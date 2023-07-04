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

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.73.1"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "4.73.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.22.0"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = "1.14.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.9.1"
    }
  }
}

# Configure GCP project id
provider "google" {
  project = var.project_id
}

# Configure GCP beta project id
provider "google-beta" {
  project = var.project_id
}

provider "kubectl" {}

locals {
  buyer_app_url     = "http://10.1.0.2:8000"
  seller_system_url = "https://seller-systems-service.com"
  registry_url      = "https://preprod.registry.ondc.org/ondc"
  gateway_url       = "https://preprod.gateway.ondc.org"

  subscriber_id    = "example.com"
  request_id       = "484be40a-3806-475d-a168-a6ec03d7b310"
  key_id           = "ec7ae8e8-f211-40ac-946f-a3407d0a76bb"
  ondc_environment = "pre-production"

  location = "us-central1"
}

locals {
  artifact_registry = {
    project_id = var.project_id
    location   = var.location
    repository = "ondc-open-commerce"
  }
}

# Enable Anthos Service Mesh (required)
# Create only once
resource "google_gke_hub_feature" "servicemesh" {
  provider = google-beta

  location = "global"
  name     = "servicemesh"
}

# Create service accounts
# Sample Buyer service account
resource "google_service_account" "dev_buyer_cluster" {
  provider = google

  account_id   = "dev-buyer-cluster"
  display_name = "Dev Buyer Service Account"
}

# Sample Seller service account
resource "google_service_account" "dev_seller_cluster" {
  provider = google

  account_id   = "dev-seller-cluster"
  display_name = "Dev Seller Service Account"
}

module "dev_key_rotation" {
  source = "../../modules/key-rotation"

  project_id        = var.project_id
  artifact_registry = local.artifact_registry

  prefix    = "dev-"
  secret_id = "key-rotation-keys"
  service_accounts = [
    google_service_account.dev_buyer_cluster.email,
    google_service_account.dev_seller_cluster.email,
  ]

  registry_url  = local.registry_url
  subscriber_id = local.subscriber_id
  request_id    = local.request_id
  location      = local.location
}

module "dev_onboarding" {
  source = "../../modules/onboarding"

  project_id        = var.project_id
  prefix            = "dev-"
  artifact_registry = local.artifact_registry

  secret_id                = module.dev_key_rotation.secret_id
  request_id               = local.request_id
  registry_encrypt_pub_key = "MCowBQYDK2VuAyEAa9Wbpvd9SsrpOZFcynyt/TO3x0Yrqyys4NUGIvyxX2Q="
  location                 = local.location

  depends_on = [
    module.dev_key_rotation
  ]
}

# Example: Buyer module use
module "dev_buyer_app" {
  source = "../../modules/buyer"

  env_prefix = "dev-"

  project_id      = var.project_id
  cluster_name    = "dev-buyer-cluster"
  service_account = google_service_account.dev_buyer_cluster.email

  artifact_registry = local.artifact_registry

  region                 = local.location
  zones                  = ["us-central1-c"]
  network_name           = "dev-buyer-network"
  subnet_name            = "dev-buyer-subnet"
  subnet_ip              = "10.0.0.0/18"
  ip_range_pods_name     = "dev-buyer-ip-range-pods"
  ip_range_services_name = "dev-buyer-ip-range-services"
  ip_range_pods          = "192.168.0.0/18"
  ip_range_services      = "192.168.64.0/18"
  buyer_app_allow_hosts  = ["10.0.0.7/18"]

  horizontal_pod_autoscaling = false
  node_pool_name             = "dev-buyer-node-pool"
  initial_node_count         = 3
  min_node_count             = 3
  max_node_count             = 10

  // User defined variables
  pubsub_prefix           = "dev-buyer"
  spanner_instance_name   = "dev-buyer-spanner-instance"
  spanner_database_name   = "dev-buyer-spanner-database"
  spanner_processing_unit = 100

  key_id         = local.key_id
  subscriber_id  = local.subscriber_id
  subscriber_url = "https://example.com/buyer/bap"

  secret_id = module.dev_key_rotation.secret_id

  buyer_app_url    = local.buyer_app_url
  registry_url     = local.registry_url
  gateway_url      = local.gateway_url
  ondc_environment = local.ondc_environment
}

# Example: Seller module use
module "dev_seller_app" {
  source = "../../modules/seller"

  env_prefix = "dev-"

  project_id      = var.project_id
  cluster_name    = "dev-seller-cluster"
  service_account = google_service_account.dev_seller_cluster.email

  artifact_registry = local.artifact_registry

  region                 = local.location
  zones                  = ["us-central1-c"]
  network_name           = "dev-seller-network"
  subnet_name            = "dev-seller-subnet"
  subnet_ip              = "10.0.0.0/18"
  ip_range_pods_name     = "dev-seller-ip-range-pods"
  ip_range_services_name = "dev-seller-ip-range-services"
  ip_range_pods          = "192.168.0.0/18"
  ip_range_services      = "192.168.64.0/18"

  horizontal_pod_autoscaling = false
  node_pool_name             = "dev-seller-node-pool"
  initial_node_count         = 3
  min_node_count             = 3
  max_node_count             = 10

  // User defined variables
  pubsub_prefix           = "dev-seller"
  spanner_instance_name   = "dev-seller-spanner-instance"
  spanner_database_name   = "dev-seller-spanner-database"
  spanner_processing_unit = 100

  key_id         = local.key_id
  subscriber_id  = local.subscriber_id
  subscriber_url = "https://example.com/seller/bpp"

  secret_id = module.dev_key_rotation.secret_id

  seller_system_url = local.seller_system_url
  registry_url      = local.registry_url
  gateway_url       = local.gateway_url
  ondc_environment  = local.ondc_environment
}

module "dev-loadbalancer" {
  source = "../../modules/helpers/loadbalancer"

  env_prefix = "dev-"
  name       = "example-lb"

  address                   = "122.96.84.212"
  domains                   = ["example.com"]
  random_certificate_suffix = true

  buyer = {
    backend_name            = "dev-buyer-backend"
    cluster_name            = module.dev_buyer_app.cluster_name
    network_name            = module.dev_buyer_app.network_name
    network_endpoint_groups = module.dev_buyer_app.neg
    max_rate_per_endpoint   = 10
  }

  seller = {
    backend_name            = "dev-seller-backend"
    cluster_name            = module.dev_seller_app.cluster_name
    network_name            = module.dev_seller_app.network_name
    network_endpoint_groups = module.dev_seller_app.neg
    max_rate_per_endpoint   = 10
  }

  onboarding = {
    backend_name   = "dev-onboarding"
    serverless_neg = module.dev_onboarding.serverless_neg
  }
}
