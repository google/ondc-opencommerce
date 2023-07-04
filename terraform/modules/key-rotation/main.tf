# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

locals {
  key_rotater = {
    account_id   = "${var.prefix}key-rotater"
    display_name = "Key Rotater"
  }
  rotation_triggerer = {
    account_id   = "${var.prefix}rotation-trigger"
    display_name = "Rotation Trigger"
  }
}

locals {
  artifact_registry = {
    project_id = var.artifact_registry.project_id
    location   = var.artifact_registry.location
    repository = var.artifact_registry.repository
  }
}

# Enable Cloud Pub/Sub API
resource "google_project_service" "pubsub" {
  provider = google

  service = "pubsub.googleapis.com"

  disable_dependent_services = false
  disable_on_destroy         = false
}

// SECRET MANAGER ADMIN
resource "google_secret_manager_secret_iam_member" "secretmanagerAdmin" {
  provider = google
  for_each = {
    for idx, service_account in var.service_accounts : idx => "serviceAccount:${service_account}"
  }

  project   = var.project_id
  secret_id = google_secret_manager_secret.keys.id
  role      = "roles/secretmanager.admin"
  member    = each.value
}

// Enable Secret Manager API
resource "google_project_service" "secret_manager" {
  provider = google

  service = "secretmanager.googleapis.com"

  disable_dependent_services = false
  disable_on_destroy         = false
}

// Enable Cloud Run API
resource "google_project_service" "cloud_run" {
  provider = google

  service = "run.googleapis.com"

  disable_dependent_services = false
  disable_on_destroy         = false
}

// Create Service Account (key rotator)
resource "google_service_account" "key_rotater" {
  provider = google

  account_id   = local.key_rotater.account_id
  display_name = local.key_rotater.display_name
}

// Secret Manager IAM Member (secret version adder)
resource "google_secret_manager_secret_iam_member" "key_rotater_secret_adder" {
  provider = google

  secret_id = google_secret_manager_secret.keys.id
  role      = "roles/secretmanager.secretVersionAdder"
  member    = "serviceAccount:${google_service_account.key_rotater.email}"
}

// Create Service Account (rotation triggerer)
resource "google_service_account" "rotation_trigger" {
  provider = google

  account_id   = local.rotation_triggerer.account_id
  display_name = local.rotation_triggerer.display_name
}

// Create Pub/Sub Agent
resource "google_project_service_identity" "pubsub_agent" {
  provider = google-beta

  service = "pubsub.googleapis.com"
}

// Add service accont token creator role to pubsub agent
resource "google_project_iam_member" "project_token_creator" {
  provider = google

  project = var.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${google_project_service_identity.pubsub_agent.email}"
}

// Create cloud run service (key rotation)
resource "google_cloud_run_service" "key_rotater" {
  provider = google

  name     = "${var.prefix}key-rotater"
  location = var.location

  template {
    metadata {
      annotations = {
        "autoscaling.knative.dev/minScale" = 0
        "autoscaling.knative.dev/maxScale" = 1
      }
    }

    spec {
      service_account_name = google_service_account.key_rotater.email
      containers {
        image = "${local.artifact_registry.location}-docker.pkg.dev/${local.artifact_registry.project_id}/${local.artifact_registry.repository}/key-rotation:latest"
        ports {
          container_port = 8080
        }
        dynamic "env" {
          for_each = {
            "PROJECT_ID"      = var.project_id,
            "SECRET_ID"       = google_secret_manager_secret.keys.id,
            "REGISTRY_URL"    = var.registry_url
            "REQUEST_ID"      = var.request_id
            "SUBSCRIBER_ID"   = var.subscriber_id
            "ROTATION_PERIOD" = var.rotation_period
          }
          content {
            name  = env.key
            value = env.value
          }
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  metadata {
    annotations = {
      "run.googleapis.com/ingress" = "internal"
    }
  }

  depends_on = [google_project_service.cloud_run]
}

// Add Invoker role to rotation triggerer
resource "google_cloud_run_service_iam_member" "rotation_trigger_run_invoker" {
  provider = google

  service  = google_cloud_run_service.key_rotater.name
  location = google_cloud_run_service.key_rotater.location
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.rotation_trigger.email}"
}

// Generate random topic suffix
resource "random_id" "topic_suffix" {
  byte_length = 4
}

// Create Pub/Sub topic (key rotation) 
resource "google_pubsub_topic" "key_rotation" {
  provider = google

  name = "${var.prefix}key-rotation-${random_id.topic_suffix.hex}"

  depends_on = [google_project_service.pubsub]
}

// Generate random subscription suffix
resource "random_id" "subscription_suffix" {
  byte_length = 4
}

// Create Pub/Sub subscriptions 
resource "google_pubsub_subscription" "key_rotation" {
  provider = google

  name  = "${var.prefix}key-rotation-${random_id.subscription_suffix.hex}"
  topic = google_pubsub_topic.key_rotation.name

  push_config {
    push_endpoint = google_cloud_run_service.key_rotater.status[0].url

    oidc_token {
      service_account_email = google_service_account.rotation_trigger.email
    }
  }

  // prevent Key Rotation service to retry too many times
  message_retention_duration = "900s" // 15 min

  retry_policy {
    minimum_backoff = "300s" // 5 min
  }

  depends_on = [
    google_cloud_run_service_iam_member.rotation_trigger_run_invoker,
  ]
}

// Store all keys in one secret
resource "google_secret_manager_secret" "keys" {
  provider = google

  secret_id = var.secret_id

  replication {
    automatic = true
  }

  topics {
    name = google_pubsub_topic.key_rotation.id
  }

  rotation {
    rotation_period    = var.rotation_period // 6 months duraton in second
    next_rotation_time = timeadd(timestamp(), "7m")
  }

  depends_on = [
    google_project_service.secret_manager,
    google_pubsub_topic_iam_member.sm_sa_publisher
  ]

  lifecycle {
    ignore_changes = [
      rotation[0].next_rotation_time
    ]
  }
}

resource "google_project_service_identity" "secret_manager_identity" {
  provider = google-beta

  service = "secretmanager.googleapis.com"
}

resource "google_pubsub_topic_iam_member" "sm_sa_publisher" {
  provider = google

  role   = "roles/pubsub.publisher"
  member = "serviceAccount:${google_project_service_identity.secret_manager_identity.email}"
  topic  = google_pubsub_topic.key_rotation.name
}

