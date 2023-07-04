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

// --- REQUIRED --- //
variable "cluster_name" {
  type        = string
  description = "GKE Cluster Name"
}

variable "service_account" {
  type        = string
  description = "GKE Cluster Service Account"
}

variable "pubsub_prefix" {
  type        = string
  description = "Prefix of each Pub/Sub resource"
  default     = "seller"
}

variable "spanner_instance_name" {
  type        = string
  description = "Spanner Instance name"
  default     = "seller-ondc-spanner-instance"
}

variable "spanner_processing_unit" {
  type        = number
  description = "Spanner Processing Unit"
}

variable "spanner_database_name" {
  type        = string
  description = "Spanner Database name"
  default     = "seller-ondc-spanner-database"
}

variable "key_id" {
  type        = string
  description = "Unique Key ID of our entity that is registered to the ONDC network"
}

variable "subscriber_id" {
  type        = string
  description = "Subscriber ID of the entity in the ONDC network eg. `abcstore.com`"
}

variable "subscriber_url" {
  type        = string
  description = "Subscriber URL of the entity in the ONDC network eg. `https://abcstore.com/bpp`"
}

variable "secret_id" {
  type        = string
  description = "Secret Manager's Secret ID that store our key pairs"
}

variable "seller_system_url" {
  type        = string
  description = "Seller System's URL for receiving seller request eg. /search"
}

variable "registry_url" {
  type        = string
  description = "ONDC Registry URL" // TODO: add a clearer description
}

variable "gateway_url" {
  type        = string
  description = "ONDC Gateway URL"
}

variable "artifact_registry" {
  type = object({
    project_id = string,
    location   = string,
    repository = string,
  })
  description = "Artifact Registry where the Docker images stored"
}

// --- OPTIONAL --- //
variable "project_id" {
  type        = string
  description = "Google Cloud Project ID"
  default     = ""
}

variable "env_prefix" {
  type        = string
  description = "Environment Prefix. This will be use as a prefix of resources that cannot be duplicated."
  default     = ""
}

variable "ondc_environment" {
  type        = string
  description = "Network environment of ONDC. It should be one of staging, pre-production, production"
  default     = "staging"
}

variable "spanner_display_name" {
  type        = string
  description = "Spanner Instance Display Name"
  default     = "Seller Spanner Instance"
}

variable "network_name" {
  type        = string
  description = "GKE Network Name"
  default     = ""
}

variable "subnet_name" {
  type        = string
  description = "GKE Subnet name"
  default     = ""
}

variable "subnet_ip" {
  type        = string
  description = "GKE Node IP Range"
  default     = ""
}

variable "ip_range_pods_name" {
  type        = string
  description = "GKE Pod IP Range's Name. Default: {cluster_name}-ip-range-pods"
  default     = ""
}
variable "ip_range_services_name" {
  type        = string
  description = "GKE Service IP Range's Name. Default: {cluster_name}-ip-range-services"
  default     = ""
}
variable "ip_range_pods" {
  type        = string
  description = "GKE Pod IP Range"
  default     = "192.168.0.0/18"
}
variable "ip_range_services" {
  type        = string
  description = "GKE Service IP Range"
  default     = "192.168.64.0/18"
}
variable "region" {
  type        = string
  description = "GKE Network Region"
  default     = "us-central1"
}

variable "zones" {
  type        = list(string)
  description = "GKE Network Zones"
  default     = ["us-central1-c"]
}

variable "node_pool_name" {
  type        = string
  description = "GKE Node Pool Name"
  default     = "default-node-pool"
}

variable "initial_node_count" {
  type        = number
  description = "Initial Number of Node within the Node Pool"
  default     = 10
}

variable "max_node_count" {
  type        = number
  description = "Maximum Number of Node within the Node Pool"
  default     = 100
}

variable "min_node_count" {
  type        = number
  description = "Minimum Number of Node within the Node Pool"
  default     = 5
}

variable "machine_type" {
  type        = string
  description = "Machine type of VM in the cluster. Refer to https://cloud.google.com/service-mesh/docs/unified-install/anthos-service-mesh-prerequisites#cluster_requirements for details."
  default     = "e2-standard-4"

}

variable "allow_hosts" {
  type        = list(string)
  description = "List of Allowed Hosts"
  default     = ["*"]
}

variable "horizontal_pod_autoscaling" {
  type        = bool
  description = "Enable Auto Pods Scaling"
  default     = true
}
