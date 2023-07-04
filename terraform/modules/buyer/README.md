# ONDC Buyer Module

Terraform module to deploy Core API adapter for buyer app.

## Features

Key components of this module consist of
- GKE Cluster with the following services deployed in the cluster
    * Buyer App Service
    * Request Action Service
    * BAP API
    * BAP Adapter Service
- Pub/Sub topics and subscriptions for relaying messages.
- Spanner Database for storing transactions
  To see database schema, see [Spanner module](../internal/spanner/)

## Communication flow
This is an overview of how messasges are being relayed through this module.
1. Your open-commerce buyer application sends a message (eg. /search) to the Buyer App Service. The Buyer App Service validates the JSON payload.
2. The Buyer App Service publish the message to the Pub/Sub topic.
3. The Request Action Service pulls the message from the Pub/Sub topic, creates an auth header and sends it to the ONDC network.
4. The BAP API receives a callback message (eg. /on_search), the service validates an auth header and JSON payload, and publishes it to a Pub/Sub topic.
5. The BAP Adapter Service pulls the message from the Pub/Sub topic and sends it to your open-commerce buyer application.

## Requirements for connecting to Buyer Module
1. Open-commerce buyer application that implements ONDC buyer API.
2. Setting up Ingress and Egress of the services
   - connect Buyer App Service with your open-commerce buyer application
   - expose BAP API to the internet.
   You are free to design and add the required networking components as needed. We provide [Load Balancer module](../helpers/loadbalancer/) as an helper module that you can use.

## Example Usage
See the [terraform/examples/sample](../../examples/sample/)

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_google"></a> [google](#requirement\_google) | 4.73.1 |
| <a name="requirement_google-beta"></a> [google-beta](#requirement\_google-beta) | 4.73.1 |
| <a name="requirement_kubectl"></a> [kubectl](#requirement\_kubectl) | 1.14.0 |
| <a name="requirement_random"></a> [random](#requirement\_random) | 3.5.1 |
| <a name="requirement_time"></a> [time](#requirement\_time) | 0.9.1 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | 4.73.1 |
| <a name="provider_google-beta"></a> [google-beta](#provider\_google-beta) | 4.73.1 |
| <a name="provider_kubectl"></a> [kubectl](#provider\_kubectl) | 1.14.0 |
| <a name="provider_random"></a> [random](#provider\_random) | 3.5.1 |
| <a name="provider_time"></a> [time](#provider\_time) | 0.9.1 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_artifact_registry"></a> [artifact\_registry](#input\_artifact\_registry) | Artifact Registry where the Docker images stored | <pre>object({<br>    project_id = string,<br>    location   = string,<br>    repository = string,<br>  })</pre> | n/a | yes |
| <a name="input_buyer_app_allow_hosts"></a> [buyer\_app\_allow\_hosts](#input\_buyer\_app\_allow\_hosts) | List of Allowed Hosts for Buyer App service. Default to `*` so that it can be called from the internet. It can be changed to IP address of the istio ingress gateway to secure the endpoint. | `list(string)` | <pre>[<br>  "*"<br>]</pre> | no |
| <a name="input_buyer_app_url"></a> [buyer\_app\_url](#input\_buyer\_app\_url) | Buyer Application's URL for receiving buyer request eg. /on\_search | `string` | n/a | yes |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | GKE Cluster Name | `string` | n/a | yes |
| <a name="input_env_prefix"></a> [env\_prefix](#input\_env\_prefix) | Environment Prefix. This will be use as a prefix of resources that cannot be duplicated. | `string` | `""` | no |
| <a name="input_gateway_url"></a> [gateway\_url](#input\_gateway\_url) | ONDC Gateway URL | `string` | n/a | yes |
| <a name="input_horizontal_pod_autoscaling"></a> [horizontal\_pod\_autoscaling](#input\_horizontal\_pod\_autoscaling) | Enable Auto Pods Scaling | `bool` | `true` | no |
| <a name="input_initial_node_count"></a> [initial\_node\_count](#input\_initial\_node\_count) | Initial Number of Node within the Node Pool | `number` | `10` | no |
| <a name="input_ip_range_pods"></a> [ip\_range\_pods](#input\_ip\_range\_pods) | GKE Pod IP Range | `string` | `"192.168.0.0/18"` | no |
| <a name="input_ip_range_pods_name"></a> [ip\_range\_pods\_name](#input\_ip\_range\_pods\_name) | GKE Pod IP Range's Name. Default: {cluster\_name}-ip-range-pods | `string` | `""` | no |
| <a name="input_ip_range_services"></a> [ip\_range\_services](#input\_ip\_range\_services) | GKE Service IP Range | `string` | `"192.168.64.0/18"` | no |
| <a name="input_ip_range_services_name"></a> [ip\_range\_services\_name](#input\_ip\_range\_services\_name) | GKE Service IP Range's Name. Default: {cluster\_name}-ip-range-services | `string` | `""` | no |
| <a name="input_key_id"></a> [key\_id](#input\_key\_id) | Unique Key ID of our entity that is registered to the ONDC network | `string` | n/a | yes |
| <a name="input_machine_type"></a> [machine\_type](#input\_machine\_type) | Machine type of VM in the cluster. Refer to https://cloud.google.com/service-mesh/docs/unified-install/anthos-service-mesh-prerequisites#cluster_requirements for details. | `string` | `"e2-standard-4"` | no |
| <a name="input_max_node_count"></a> [max\_node\_count](#input\_max\_node\_count) | Maximum Number of Node within the Node Pool | `number` | `100` | no |
| <a name="input_min_node_count"></a> [min\_node\_count](#input\_min\_node\_count) | Minimum Number of Node within the Node Pool | `number` | `5` | no |
| <a name="input_network_name"></a> [network\_name](#input\_network\_name) | GKE Network Name | `string` | `""` | no |
| <a name="input_node_pool_name"></a> [node\_pool\_name](#input\_node\_pool\_name) | GKE Node Pool Name | `string` | `"default-node-pool"` | no |
| <a name="input_ondc_environment"></a> [ondc\_environment](#input\_ondc\_environment) | Network environment of ONDC. It should be one of staging, pre-production, production | `string` | `"staging"` | no |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | Google Cloud Project ID | `string` | `""` | no |
| <a name="input_pubsub_prefix"></a> [pubsub\_prefix](#input\_pubsub\_prefix) | Prefix of each Pub/Sub resource | `string` | `"buyer"` | no |
| <a name="input_region"></a> [region](#input\_region) | GKE Network Region | `string` | `"us-central1"` | no |
| <a name="input_registry_url"></a> [registry\_url](#input\_registry\_url) | ONDC Registry URL | `string` | n/a | yes |
| <a name="input_secret_id"></a> [secret\_id](#input\_secret\_id) | Secret Manager's Secret ID that store our key pairs | `string` | n/a | yes |
| <a name="input_service_account"></a> [service\_account](#input\_service\_account) | GKE Cluster Service Account | `string` | n/a | yes |
| <a name="input_spanner_database_name"></a> [spanner\_database\_name](#input\_spanner\_database\_name) | Spanner Database name | `string` | `"buyer-ondc-spanner-database"` | no |
| <a name="input_spanner_display_name"></a> [spanner\_display\_name](#input\_spanner\_display\_name) | Spanner Instance Display Name | `string` | `"Buyer Spanner Instance"` | no |
| <a name="input_spanner_instance_name"></a> [spanner\_instance\_name](#input\_spanner\_instance\_name) | Spanner Instance name | `string` | `"buyer-ondc-spanner-instance"` | no |
| <a name="input_spanner_processing_unit"></a> [spanner\_processing\_unit](#input\_spanner\_processing\_unit) | Spanner Processing Unit | `number` | n/a | yes |
| <a name="input_subnet_ip"></a> [subnet\_ip](#input\_subnet\_ip) | GKE Node IP Range | `string` | `""` | no |
| <a name="input_subnet_name"></a> [subnet\_name](#input\_subnet\_name) | GKE Subnet name | `string` | `""` | no |
| <a name="input_subscriber_id"></a> [subscriber\_id](#input\_subscriber\_id) | Subscriber ID of the entity in the ONDC network eg. `abcstore.com` | `string` | n/a | yes |
| <a name="input_subscriber_url"></a> [subscriber\_url](#input\_subscriber\_url) | Subscriber URL of the entity in the ONDC network eg. `https://abcstore.com/bap` | `string` | n/a | yes |
| <a name="input_zones"></a> [zones](#input\_zones) | GKE Network Zones | `list(string)` | <pre>[<br>  "us-central1-c"<br>]</pre> | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | GKE Cluster Name |
| <a name="output_neg"></a> [neg](#output\_neg) | Network Endpoint Groups |
| <a name="output_network_name"></a> [network\_name](#output\_network\_name) | GKE Network Name |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_gke"></a> [gke](#module\_gke) | terraform-google-modules/kubernetes-engine/google | 27.0.0 |
| <a name="module_network"></a> [network](#module\_network) | terraform-google-modules/network/google | 7.1.0 |
| <a name="module_pubsub"></a> [pubsub](#module\_pubsub) | ../internal/pubsub | n/a |
| <a name="module_spanner"></a> [spanner](#module\_spanner) | ../internal/spanner | n/a |

## Resources

| Name | Type |
|------|------|
| [google-beta_google_gke_hub_feature_membership.feature_member](https://registry.terraform.io/providers/hashicorp/google-beta/4.73.1/docs/resources/google_gke_hub_feature_membership) | resource |
| [google-beta_google_gke_hub_membership.fleet_membership](https://registry.terraform.io/providers/hashicorp/google-beta/4.73.1/docs/resources/google_gke_hub_membership) | resource |
| [google-beta_google_project_iam_member.hub_service_agent_gke](https://registry.terraform.io/providers/hashicorp/google-beta/4.73.1/docs/resources/google_project_iam_member) | resource |
| [google-beta_google_project_iam_member.workload_identity_user](https://registry.terraform.io/providers/hashicorp/google-beta/4.73.1/docs/resources/google_project_iam_member) | resource |
| [google-beta_google_project_service_identity.sa_gkehub](https://registry.terraform.io/providers/hashicorp/google-beta/4.73.1/docs/resources/google_project_service_identity) | resource |
| [google_artifact_registry_repository_iam_member.reader](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/artifact_registry_repository_iam_member) | resource |
| [google_compute_firewall.allow_healthcheck_and_proxy](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/compute_firewall) | resource |
| [google_project_iam_member.k8s_member_workload_identity](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/project_iam_member) | resource |
| [google_project_iam_member.logWriter](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/project_iam_member) | resource |
| [google_project_iam_member.metricWriter](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/project_iam_member) | resource |
| [google_project_iam_member.publisher](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/project_iam_member) | resource |
| [google_project_iam_member.pubsubAdmin](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/project_iam_member) | resource |
| [google_project_iam_member.viewer](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/project_iam_member) | resource |
| [google_spanner_database_iam_member.spannerDatabaseAdmin](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/spanner_database_iam_member) | resource |
| [kubectl_manifest.allow_egress_googleapis](https://registry.terraform.io/providers/gavinbunney/kubectl/1.14.0/docs/resources/manifest) | resource |
| [kubectl_manifest.app_deployments](https://registry.terraform.io/providers/gavinbunney/kubectl/1.14.0/docs/resources/manifest) | resource |
| [kubectl_manifest.app_gateways](https://registry.terraform.io/providers/gavinbunney/kubectl/1.14.0/docs/resources/manifest) | resource |
| [kubectl_manifest.app_services](https://registry.terraform.io/providers/gavinbunney/kubectl/1.14.0/docs/resources/manifest) | resource |
| [kubectl_manifest.app_servieaccounts](https://registry.terraform.io/providers/gavinbunney/kubectl/1.14.0/docs/resources/manifest) | resource |
| [kubectl_manifest.configs](https://registry.terraform.io/providers/gavinbunney/kubectl/1.14.0/docs/resources/manifest) | resource |
| [kubectl_manifest.istio_ingress](https://registry.terraform.io/providers/gavinbunney/kubectl/1.14.0/docs/resources/manifest) | resource |
| [kubectl_manifest.namespaces](https://registry.terraform.io/providers/gavinbunney/kubectl/1.14.0/docs/resources/manifest) | resource |
| [random_id.suffix](https://registry.terraform.io/providers/hashicorp/random/3.5.1/docs/resources/id) | resource |
| [time_sleep.for_asm_ready](https://registry.terraform.io/providers/hashicorp/time/0.9.1/docs/resources/sleep) | resource |
| [google_client_config.main](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/data-sources/client_config) | data source |
| [google_compute_network_endpoint_group.neg](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/data-sources/compute_network_endpoint_group) | data source |
| [google_project.main](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/data-sources/project) | data source |

