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

// REQUIRED VARIABLES //

variable "project_id" {
  type        = string
  description = "Google Cloud Project ID"
}

variable "secret_id" {
  type        = string
  description = "Secret Manager's Secret ID"
}

variable "service_accounts" {
  type        = list(string)
  description = "Service Accounts List as Secret Manager Admins"
}

variable "registry_url" {
  type        = string
  description = "ONDC Registry URL"
}

variable "request_id" {
  type        = string
  description = "Arbitary ID (eg. UUID). This will be used when sending key rotation request to ONDC registry. It should be the same ID you will use in `onboarding` module."
}

variable "subscriber_id" {
  type        = string
  description = "Subscriber ID of the ONDC entity ex. `ondcaccelerator.com`"
}

variable "location" {
  type        = string
  description = "Cloud Run location."
}

variable "artifact_registry" {
  type = object({
    project_id = string,
    location   = string,
    repository = string,
  })
  description = "Artifact Registry where the Docker images stored"
}

// OPTIONAL VARIABLES //

variable "rotation_period" {
  type        = string
  description = "Time between each key rotation. Default to 6 months. **WARNING**: changing this field after created the Secret Manager secret can delete all sercet versions. See this [issue](https://github.com/hashicorp/terraform-provider-google/issues/13770)"
  default     = "15780000s" # 6 months duration
}

variable "prefix" {
  type        = string
  description = "Resouce Prefix. If it's not empty, it should contains `-` as a last character eg. `dev-`"
  default     = ""
}

