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
  instance_name   = var.instance_name == "" ? "ondc-instance" : var.instance_name
  database_name   = var.database_name == "" ? "ondc-database" : var.database_name
  instance_config = var.instance_config == "" ? "regional-us-central1" : var.instance_config
  force_destroy   = var.force_destroy
  display_name    = var.display_name == "" ? title(replace(local.instance_name, "-", " ")) : var.display_name
}

// Enable spanner API
resource "google_project_service" "spanner" {
  provider = google

  service = "spanner.googleapis.com"

  disable_dependent_services = false
  disable_on_destroy         = false
}

// Create spanner instance
resource "google_spanner_instance" "spanner_instance" {
  provider = google

  name             = local.instance_name
  display_name     = local.display_name
  config           = local.instance_config
  processing_units = var.processing_unit
  depends_on       = [google_project_service.spanner]

  force_destroy = local.force_destroy
}

# This is a hack to remove license comment from the ddl string.
# Since it is an error to send a ddl statement with comments.
locals {
  registration_ddl = split("\n\n", file("${path.module}/sql/registration_table.sql"))[1]
  transaction_ddl  = split("\n\n", file("${path.module}/sql/transaction_table.sql"))[1]
}

// Create spanner database
resource "google_spanner_database" "spanner_database" {
  provider = google

  instance = google_spanner_instance.spanner_instance.name
  name     = local.database_name
  ddl = [
    local.registration_ddl,
    local.transaction_ddl
  ]
}
