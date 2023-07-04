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

locals {
  project_id = var.project_id == "" ? data.google_client_config.main.project : var.project_id
  env_prefix = var.env_prefix
  name       = var.name
  backends   = [var.buyer, var.seller]
}

# Create URL MAPs
resource "google_compute_url_map" "default" {
  name = "${local.env_prefix}${local.name}"

  default_service = module.http_lb.backend_services[local.backends[0].backend_name].self_link
  host_rule {
    hosts        = ["*"]
    path_matcher = "allpaths"
  }
  path_matcher {
    name            = "allpaths"
    default_service = module.http_lb.backend_services[local.backends[0].backend_name].self_link

    dynamic "path_rule" {
      for_each = var.buyer == null ? [] : [1]
      content {
        paths   = ["/buyer/*"]
        service = module.http_lb.backend_services[lookup(var.buyer, "backend_name", "buyer-backend")].self_link
      }
    }

    dynamic "path_rule" {
      for_each = var.seller == null ? [] : [1]
      content {
        paths   = ["/seller/*"]
        service = module.http_lb.backend_services[lookup(var.seller, "backend_name", "seller-backend")].self_link
      }
    }

    dynamic "path_rule" {
      for_each = var.onboarding == null ? [] : [1]
      content {
        paths   = ["/onboarding/*"]
        service = one(google_compute_backend_service.onboarding[*].self_link)
        route_action {
          url_rewrite {
            path_prefix_rewrite = "/"
          }
        }
      }
    }

    dynamic "path_rule" {
      for_each = var.onboarding == null ? [] : [1]
      content {
        paths   = ["/ondc-site-verification.html"]
        service = one(google_compute_backend_service.onboarding[*].self_link)
      }
    }
  }
}

resource "google_compute_backend_service" "onboarding" {
  count      = var.onboarding == null ? 0 : 1
  name       = "${local.env_prefix}onboarding-backend"
  protocol   = "HTTP"
  port_name  = "http"
  enable_cdn = false

  backend {
    group = var.onboarding.serverless_neg.id
  }
  log_config {
    enable      = true
    sample_rate = 1.0
  }
}

locals {
  health_check = {
    request_path = "/healthz/ready"
    port         = "15021"
  }
}

module "http_lb" {
  source  = "GoogleCloudPlatform/lb-http/google"
  version = "~> 9.0"

  address        = var.address == "" ? null : var.address
  create_address = var.address == "" ? true : false

  project = local.project_id
  name    = "${local.env_prefix}${local.name}"

  ssl                             = true
  managed_ssl_certificate_domains = var.domains
  use_ssl_certificates            = false
  random_certificate_suffix       = var.random_certificate_suffix

  url_map        = google_compute_url_map.default.self_link
  create_url_map = false
  https_redirect = false
  http_forward   = false

  firewall_projects = [for backend in local.backends : local.project_id if backend != null]
  firewall_networks = [for backend in local.backends : backend.network_name if backend != null]
  target_tags       = [for backend in local.backends : "gke-${backend.cluster_name}" if backend != null]

  backends = {
    for backend in local.backends : backend.backend_name => {
      protocol  = "HTTP"
      port      = 80
      port_name = "http"

      enable_cdn = false

      health_check = local.health_check

      groups = [
        for neg in backend.network_endpoint_groups : {
          group                 = neg.id
          balancing_mode        = "RATE"
          max_rate_per_endpoint = backend.max_rate_per_endpoint
        }
      ]

      log_config = {
        enable      = true
        sample_rate = 1.0
      }

      iap_config = {
        enable = false
      }

    } if backend != null
  }
}
