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
  description = "Secret Manager's Secret ID that store our key pairs"
}

variable "request_id" {
  type        = string
  description = "Request ID (ex. uuid). This should be the same ID you will use for `/subscribe` API."
}

variable "artifact_registry" {
  type = object({
    project_id = string,
    location   = string,
    repository = string,
  })
  description = "Artifact Registry where the Docker images stored"
}

variable "registry_encrypt_pub_key" {
  type        = string
  description = "Encryption public key of the ONDC registry. This info should be avalable in the [ONDC onboarding document](https://github.com/ONDC-Official/developer-docs/blob/main/registry/Onboarding%20of%20Participants.md)"
}

variable "location" {
  type        = string
  description = "Cloud Run Location of onboarding service"
}

// OPTIONAL VARIABLES //
variable "prefix" {
  type        = string
  description = "Resouce Prefix. If it's not empty, it should contains `-` as a last character eg. `dev-`"
  default     = ""
}

