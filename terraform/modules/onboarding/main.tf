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
  secret_id               = var.secret_id
  request_id              = var.request_id
  onboarding_account_id   = "${var.prefix}onboarding-service-account"
  onboarding_display_name = "On Boarding Service Account"

  artifact_registry = {
    project_id = var.artifact_registry.project_id
    location   = var.artifact_registry.location
    repository = var.artifact_registry.repository
  }
}

resource "google_service_account" "onboarding" {
  provider = google

  account_id   = local.onboarding_account_id
  display_name = local.onboarding_display_name
}

resource "google_project_service" "cloud_run" {
  provider = google

  service = "run.googleapis.com"

  disable_dependent_services = false
  disable_on_destroy         = false
}

resource "google_project_service" "secret_manager" {
  provider = google

  service = "secretmanager.googleapis.com"

  disable_dependent_services = false
  disable_on_destroy         = false
}

resource "google_cloud_run_service_iam_member" "invoker" {
  provider = google

  service  = google_cloud_run_service.onboarding.name
  location = google_cloud_run_service.onboarding.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_secret_manager_secret_iam_member" "read" {
  provider = google

  project   = var.project_id
  secret_id = local.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.onboarding.email}"
}

locals {
  location = var.location
}

resource "google_cloud_run_service" "onboarding" {
  provider = google

  name     = "${var.prefix}onboarding"
  location = local.location

  template {
    metadata {
      annotations = {
        "autoscaling.knative.dev/minScale" = 0
        "autoscaling.knative.dev/maxScale" = 1
      }
    }
    spec {
      service_account_name = google_service_account.onboarding.email
      containers {
        image = "${local.artifact_registry.location}-docker.pkg.dev/${local.artifact_registry.project_id}/${local.artifact_registry.repository}/onboarding:latest"
        ports {
          container_port = 8080
        }
        dynamic "env" {
          for_each = {
            "PROJECT_ID"               = var.project_id,
            "SECRET_ID"                = local.secret_id,
            "REQUEST_ID"               = local.request_id
            "REGISTRY_ENCRYPT_PUB_KEY" = var.registry_encrypt_pub_key
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

  depends_on = [
    google_project_service.cloud_run
  ]
}

resource "google_compute_region_network_endpoint_group" "serverless_neg" {
  provider = google-beta

  name                  = "${var.prefix}onboarding"
  network_endpoint_type = "SERVERLESS"
  region                = local.location
  cloud_run {
    service = google_cloud_run_service.onboarding.name
  }
}
