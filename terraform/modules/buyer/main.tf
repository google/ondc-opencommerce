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

data "google_client_config" "main" {}
data "google_project" "main" {}

resource "random_id" "suffix" {
  byte_length = 4
}

locals {
  project_id = var.project_id == "" ? data.google_client_config.main.project : var.project_id
}

// User configure variables
locals {
  env_prefix    = var.env_prefix
  key_id        = var.key_id
  subscriber_id = var.subscriber_id
  secret_id     = var.secret_id
  buyer_app_url = var.buyer_app_url
  registry_url  = var.registry_url
  gateway_url   = var.gateway_url

  pubsub_prefix         = var.pubsub_prefix
  spanner_instance_name = var.spanner_instance_name
  spanner_database_name = var.spanner_database_name

  registry_project_id = var.artifact_registry.project_id
  registry_location   = var.artifact_registry.location
  registry_repository = var.artifact_registry.repository
}

// --- PERMISSIONS --- //
// LOG WRITER 
resource "google_project_iam_member" "logWriter" {
  provider = google

  project = local.project_id
  role    = "roles/logging.logWriter"

  member = "serviceAccount:${var.service_account}"
}

// METRIC WRITER
resource "google_project_iam_member" "metricWriter" {
  provider = google

  project = local.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${var.service_account}"
}

// PUBSUB ADMIN
resource "google_project_iam_member" "pubsubAdmin" {
  provider = google

  project = local.project_id
  role    = "roles/pubsub.admin"
  member  = "serviceAccount:${var.service_account}"
}

// IMAGES REGISTRY
resource "google_artifact_registry_repository_iam_member" "reader" {
  provider = google

  project    = local.registry_project_id
  location   = local.registry_location
  repository = local.registry_repository
  role       = "roles/artifactregistry.reader"

  member = "serviceAccount:${var.service_account}"
}

// SPANNER DATABASE ADMIN
resource "google_spanner_database_iam_member" "spannerDatabaseAdmin" {
  provider = google

  instance = module.spanner.instance.name
  database = module.spanner.database.name

  project = local.project_id
  role    = "roles/spanner.admin"
  member  = "serviceAccount:${var.service_account}"
}

// PUBSUB PUBLISHER
resource "google_project_iam_member" "publisher" {
  provider = google

  project = local.project_id
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${var.service_account}"
}

// PUBSUB VIEWER
resource "google_project_iam_member" "viewer" {
  provider = google

  project = local.project_id
  role    = "roles/pubsub.viewer"
  member  = "serviceAccount:${var.service_account}"
}

// --- PUBSUB --- //
module "pubsub" {
  source = "../internal/pubsub"

  prefix = local.pubsub_prefix
}

// --- SPANER --- //
module "spanner" {
  source = "../internal/spanner"

  instance_name   = local.spanner_instance_name
  database_name   = local.spanner_database_name
  processing_unit = var.spanner_processing_unit
}

// --- NETWORK --- //
locals {
  network_name           = var.network_name == "" ? "${local.cluster_name}-network-${random_id.suffix.hex}" : var.network_name
  subnet_ip              = var.subnet_ip == "" ? "10.0.0.0/18" : var.subnet_ip
  region                 = var.region == "" ? "us-central1" : var.region
  zones                  = var.zones == "" ? ["us-central1"] : var.zones
  subnet_name            = var.subnet_name == "" ? "${local.cluster_name}-network-subnet" : var.subnet_name
  ip_range_pods_name     = var.ip_range_pods_name == "" ? "${local.cluster_name}-ip-range-pods" : var.ip_range_pods_name
  ip_range_services_name = var.ip_range_services_name == "" ? "${local.cluster_name}-ip-range-services" : var.ip_range_services_name
  ip_range_pods          = var.ip_range_pods == "" ? "192.168.0.0/18" : var.ip_range_pods
  ip_range_services      = var.ip_range_services == "" ? "192.168.64.0/18" : var.ip_range_services
}

module "network" {
  source  = "terraform-google-modules/network/google"
  version = "7.1.0"

  project_id   = local.project_id
  network_name = local.network_name

  subnets = [
    {
      subnet_name   = local.subnet_name
      subnet_ip     = local.subnet_ip
      subnet_region = local.region
    }
  ]

  secondary_ranges = {
    (local.subnet_name) = [
      {
        range_name    = local.ip_range_pods_name
        ip_cidr_range = local.ip_range_pods
      },
      {
        range_name    = local.ip_range_services_name
        ip_cidr_range = local.ip_range_services
      }
    ]
  }
}

// Create Firewall Rules
resource "google_compute_firewall" "allow_healthcheck_and_proxy" {
  name    = "${local.cluster_name}-fw-allow-healthcheck-and-proxy"
  network = local.network_name

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
    ports    = ["80", "443", "15021"]
  }

  target_tags = [
    "gke-${local.cluster_name}"
  ]
  source_ranges = [
    "130.211.0.0/22",
    "35.191.0.0/16"
  ]

  depends_on = [
    module.gke,
    module.network
  ]
}

// --- GOOGLE KUBERNETES ENGINE --- //
locals {
  cluster_name = var.cluster_name
}

provider "kubernetes" {
  host                   = "https://${module.gke.endpoint}"
  token                  = data.google_client_config.main.access_token
  cluster_ca_certificate = base64decode(module.gke.ca_certificate)
}

provider "kubectl" {
  host                   = "https://${module.gke.endpoint}"
  token                  = data.google_client_config.main.access_token
  cluster_ca_certificate = base64decode(module.gke.ca_certificate)
  load_config_file       = false
}

module "gke" {
  source  = "terraform-google-modules/kubernetes-engine/google"
  version = "27.0.0"

  name                   = local.cluster_name
  project_id             = local.project_id
  create_service_account = false
  service_account        = var.service_account

  regional          = true
  region            = local.region
  zones             = local.zones
  network           = module.network.network_name
  subnetwork        = module.network.subnets_names[0]
  ip_range_pods     = local.ip_range_pods_name
  ip_range_services = local.ip_range_services_name

  # Required by Anthos Service Mesh
  cluster_resource_labels = { "mesh_id" : "proj-${data.google_project.main.number}" }

  http_load_balancing        = true
  grant_registry_access      = true
  horizontal_pod_autoscaling = var.horizontal_pod_autoscaling

  # Delete the default node pool
  # However, it cannot create a cluster without any nodes, hence initial node count is 1
  initial_node_count       = 1
  remove_default_node_pool = true

  node_pools = [
    {
      name               = var.node_pool_name
      initial_node_count = var.initial_node_count
      max_node_count     = var.max_node_count
      min_node_count     = var.min_node_count
      machine_type       = "e2-standard-4" # This manchine type is required for utilise Anthos service mesh
    }
  ]
  node_pools_oauth_scopes = {
    all = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
      "https://www.googleapis.com/auth/cloud-platform",
    ]
  }

  depends_on = [
    module.network,
    module.pubsub,
    module.spanner
  ]
}

// --- GKE HUB (FLEET MEMBERSHIPS) --- //
resource "google_project_service_identity" "sa_gkehub" {
  provider = google-beta

  project = local.project_id
  service = "gkehub.googleapis.com"
}

resource "google_project_iam_member" "hub_service_agent_gke" {
  provider = google-beta

  project = local.project_id
  role    = "roles/gkehub.serviceAgent"

  member = "serviceAccount:${var.service_account}"
}
resource "google_project_iam_member" "workload_identity_user" {
  provider = google-beta

  project = local.project_id
  role    = "roles/iam.workloadIdentityUser"
  member  = "serviceAccount:${var.service_account}"
}

resource "google_gke_hub_membership" "fleet_membership" {
  provider = google-beta

  project       = local.project_id
  membership_id = "${module.gke.name}-membership"

  endpoint {
    gke_cluster {
      resource_link = "//container.googleapis.com/${module.gke.cluster_id}"
    }
  }

  authority {
    issuer = "https://container.googleapis.com/v1/${module.gke.cluster_id}"
  }

  depends_on = [
    module.gke,
  ]
}

resource "google_gke_hub_feature_membership" "feature_member" {
  provider = google-beta

  location   = "global"
  feature    = "servicemesh"
  membership = google_gke_hub_membership.fleet_membership.membership_id
  mesh {
    management = "MANAGEMENT_AUTOMATIC"
  }
}

# Wait for some time for ASM to be ready
# Required ASM before performing Istio Ingress deployment
resource "time_sleep" "for_asm_ready" {
  create_duration = "4m" // an Estimated Time

  depends_on = [
    google_gke_hub_feature_membership.feature_member
  ]
}

// --- KUBERNETES MANIFEST FILES --- //

locals {
  configs = {
    bap_adapter_config = {
      filename   = "bap-adapter-config.yaml"
      project_id = local.project_id,
      buyer_app = {
        url = local.buyer_app_url
      }
      pubsub           = module.pubsub
      ondc_environment = var.ondc_environment
    }
    bap_apis_config = {
      filename   = "bap-apis-config.yaml"
      project_id = local.project_id,
      port       = 8080
      subscriber = {
        id = local.subscriber_id
      }
      registry = {
        url = local.registry_url
      }
      pubsub           = module.pubsub
      spanner          = module.spanner
      ondc_environment = var.ondc_environment
    }
    buyer_app_config = {
      filename         = "buyer-app-config.yaml"
      project_id       = local.project_id,
      port             = 8080,
      pubsub           = module.pubsub
      ondc_environment = var.ondc_environment
    }

    request_action_config = {
      filename   = "request-action-config.yaml"
      project_id = local.project_id,
      gateway = {
        url = local.gateway_url
      }
      secret = {
        id = local.secret_id
      }
      key = {
        id = local.key_id
      }
      subscriber = {
        id  = local.subscriber_id
        url = var.subscriber_url
      }
      pubsub           = module.pubsub
      spanner          = module.spanner
      ondc_environment = var.ondc_environment
    }
  }
}

locals {
  service_accounts = [
    {
      filename      = "bap-adapter-sa.yaml"
      name          = "bap-adapter",
      account_id    = "bap-adapter-service-account"
      display_name  = "BAP Adapter Service Account",
      k8s_name      = "bap-adapter-sa",
      k8s_namespace = "bap-adapter"
    },
    {
      filename      = "bap-apis-sa.yaml"
      name          = "bap-apis",
      account_id    = "bap-apis-service-account"
      display_name  = "BAP APIs Service Account",
      k8s_name      = "bap-apis-sa",
      k8s_namespace = "bap-apis"
    },
    {
      filename      = "buyer-app-sa.yaml"
      name          = "buyer-app",
      account_id    = "buyer-app-service-account"
      display_name  = "Buyer Application Service Account",
      k8s_name      = "buyer-app-sa",
      k8s_namespace = "buyer-app"
    },
    {
      filename      = "request-action-sa.yaml"
      name          = "request-action",
      account_id    = "request-action-service-account"
      display_name  = "Request Action Service Account",
      k8s_name      = "request-action-sa",
      k8s_namespace = "request-action"
    },
  ]
}

# Grant a workload identity role to each account so as to be able to impersonate
resource "google_project_iam_member" "k8s_member_workload_identity" {
  provider = google

  for_each = { for v in local.service_accounts : v.name => v }

  project = local.project_id
  role    = "roles/iam.workloadIdentityUser"

  member = "serviceAccount:${local.project_id}.svc.id.goog[${each.value.k8s_namespace}/${local.env_prefix}${each.value.k8s_name}]"

  depends_on = [
    module.gke,
    google_gke_hub_membership.fleet_membership
  ]
}

# Create all required namespaces
# These namespaces must be exist before applying any other resources
resource "kubectl_manifest" "namespaces" {
  for_each  = fileset(path.module, "manifests/namespaces/*.yaml")
  yaml_body = file("${path.module}/${each.value}")
  wait      = true

  depends_on = [
    module.gke
  ]
}

# Create Istio Ingress Gateway
resource "kubectl_manifest" "istio_ingress" {
  for_each = fileset(path.module, "manifests/gateways/istio_ingressgateway/*.yaml")
  yaml_body = templatefile("${path.module}/${each.value}", {
    cluster_name = local.cluster_name
  })
  wait = true

  depends_on = [
    kubectl_manifest.namespaces,
    time_sleep.for_asm_ready, // ASM Ready
  ]
}

resource "kubectl_manifest" "app_services" {
  for_each  = fileset(path.module, "manifests/app/services/*.yaml")
  yaml_body = templatefile("${path.module}/${each.value}", {})

  depends_on = [
    module.pubsub,
    module.spanner,
    kubectl_manifest.namespaces,
    kubectl_manifest.istio_ingress,
  ]
}

resource "kubectl_manifest" "configs" {
  for_each        = local.configs
  yaml_body       = templatefile("${path.module}/manifests/configs/${each.value.filename}", each.value)
  force_conflicts = true
  force_new       = true

  depends_on = [
    kubectl_manifest.namespaces,
  ]
}

resource "kubectl_manifest" "allow_egress_googleapis" {
  for_each  = fileset(path.module, "manifests/app/istio-manifest/allow-egress-googleapis/*.yaml")
  yaml_body = file("${path.module}/${each.value}")

  depends_on = [
    module.pubsub,
    kubectl_manifest.namespaces,
    kubectl_manifest.istio_ingress,
    kubectl_manifest.configs,
  ]
}

resource "kubectl_manifest" "app_servieaccounts" {
  for_each = { for v in local.service_accounts : v.name => v }
  yaml_body = templatefile("${path.module}/manifests/app/serviceaccounts/sa.yaml.tpl", {
    name            = "${var.env_prefix}${each.value.k8s_name}",
    namespace       = each.value.k8s_namespace,
    service_account = var.service_account
  })

  depends_on = [
    module.pubsub,
    kubectl_manifest.namespaces,
    kubectl_manifest.istio_ingress,
    kubectl_manifest.configs,
  ]
}

resource "kubectl_manifest" "app_gateways" {
  for_each = fileset(path.module, "manifests/app/istio-manifest/app-gateways/**/*.yaml")
  yaml_body = templatefile("${path.module}/${each.value}", {
    allow_hosts = jsonencode(var.buyer_app_allow_hosts)
  })

  depends_on = [
    module.pubsub,
    kubectl_manifest.namespaces,
    kubectl_manifest.istio_ingress,
    kubectl_manifest.configs,
  ]
}

resource "kubectl_manifest" "app_deployments" {
  for_each = fileset(path.module, "manifests/app/deployments/**/*.yaml")
  yaml_body = templatefile("${path.module}/${each.value}", {
    env_prefix = local.env_prefix
  })
  force_conflicts = true
  force_new       = true

  depends_on = [
    module.pubsub,
    module.spanner,
    kubectl_manifest.app_services,
    kubectl_manifest.app_servieaccounts,
    kubectl_manifest.namespaces,
    kubectl_manifest.configs,
    kubectl_manifest.istio_ingress,
  ]
}

data "google_compute_network_endpoint_group" "neg" {
  count = length(local.zones)
  name  = "${local.cluster_name}-http"
  zone  = local.zones[count.index]

  depends_on = [
    kubectl_manifest.istio_ingress
  ]
}
