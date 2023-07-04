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
  send_subscriptions = {
    0 : { name : "send-search", filter : "attributes.action = \"search\"" },
    1 : { name : "send-select", filter : "attributes.action = \"select\"" },
    2 : { name : "send-init", filter : "attributes.action = \"init\"" },
    3 : { name : "send-confirm", filter : "attributes.action = \"confirm\"" },
    4 : { name : "send-status", filter : "attributes.action = \"status\"" },
    5 : { name : "send-track", filter : "attributes.action = \"track\"" },
    6 : { name : "send-cancel", filter : "attributes.action = \"cancel\"" },
    7 : { name : "send-update", filter : "attributes.action = \"update\"" },
    8 : { name : "send-rating", filter : "attributes.action = \"rating\"" },
    9 : { name : "send-support", filter : "attributes.action = \"support\"" },
  }
  callback_subscriptions = {
    0 : { name : "callback-on-search", filter : "attributes.action = \"on_search\"" },
    1 : { name : "callback-on-select", filter : "attributes.action = \"on_select\"" },
    2 : { name : "callback-on-init", filter : "attributes.action = \"on_init\"" },
    3 : { name : "callback-on-confirm", filter : "attributes.action = \"on_confirm\"" },
    4 : { name : "callback-on-status", filter : "attributes.action = \"on_status\"" },
    5 : { name : "callback-on-track", filter : "attributes.action = \"on_track\"" },
    6 : { name : "callback-on-cancel", filter : "attributes.action = \"on_cancel\"" },
    7 : { name : "callback-on-update", filter : "attributes.action = \"on_update\"" },
    8 : { name : "callback-on-rating", filter : "attributes.action = \"on_rating\"" },
    9 : { name : "callback-on-support", filter : "attributes.action = \"on_support\"" },
  }
}

# Enable Cloud Pub/Sub API
resource "google_project_service" "pubsub" {
  provider = google

  service = "pubsub.googleapis.com"

  disable_dependent_services = false
  disable_on_destroy         = false
}

# Create Pub/Sub topic (Send)
resource "google_pubsub_topic" "send" {
  provider = google

  name = "${var.prefix}-send"

  depends_on = [google_project_service.pubsub]
}

# Create Pub/Sub all subscriptions (send)
resource "google_pubsub_subscription" "send" {
  provider = google

  for_each = local.send_subscriptions

  name                         = "${var.prefix}-${each.value.name}"
  topic                        = google_pubsub_topic.send.name
  message_retention_duration   = var.send_message_retention_duration
  enable_exactly_once_delivery = true
  filter                       = each.value.filter
}

// Create Pub/Sub topic (callback)
resource "google_pubsub_topic" "callback" {
  provider = google

  name = "${var.prefix}-callback"

  depends_on = [google_project_service.pubsub]
}

// Create Pub/Sub all subscriptions (callback)
resource "google_pubsub_subscription" "callback" {
  provider = google

  for_each = local.callback_subscriptions

  name                         = "${var.prefix}-${each.value.name}"
  topic                        = google_pubsub_topic.callback.name
  message_retention_duration   = var.callback_message_retention_duration
  enable_exactly_once_delivery = true
  filter                       = each.value.filter
}
