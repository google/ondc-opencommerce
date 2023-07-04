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

// Project Id
variable "project_id" {
  type        = string
  description = "GCP Project ID"
  default     = ""
}

// Environment Prefix
variable "env_prefix" {
  type        = string
  description = "Environment, use as a prefix for each resource/service."
  default     = ""
}

variable "name" {
  type        = string
  description = "Load Balancer Name"
  default     = "http-lb"
}

variable "domains" {
  type        = list(string)
  description = "Managed SSL Certificate Domains"
  default     = null
}

variable "address" {
  type        = string
  description = "Load Balancer Address. If not specify, it will be created one"
  default     = ""
}

variable "random_certificate_suffix" {
  type        = bool
  description = "Bool to enable/disable random certificate name generation. Set and keep this to true if you need to change the SSL cert."
  default     = false
}

variable "buyer" {
  description = "Buyer"
  type = object({
    backend_name = optional(string)
    cluster_name = string
    network_name = string
    network_endpoint_groups = list(object({
      id = string
    }))
    max_rate_per_endpoint = number
  })
  default = null
}

variable "seller" {
  description = "Seller"
  type = object({
    backend_name = optional(string)
    cluster_name = string
    network_name = string
    network_endpoint_groups = list(object({
      id = string
    }))
    max_rate_per_endpoint = number
  })
  default = null
}

variable "onboarding" {
  description = "Onboarding"
  type = object({
    backend_name = optional(string)
    serverless_neg = object({
      id = string
    })
    paths = optional(list(string))
  })
  default = null
}
