# Onboarding Module

Terraform module to deploy the Onboarding services.

## Features
This module should be used when you are onboarding to ONDC network in pre-prod or prod environment.
If you are onboarding to the network in staging environment, this module is not neccessary.

It provides a onboarding service that implements `/on_subscribe` API and `ondc-site-verification.html` API defined in [ONDC onboarding guide](https://github.com/ONDC-Official/developer-docs/blob/main/registry/Onboarding%20of%20Participants.md). The service will be deployed on `cloud run`.


## Example Usage
See the [terraform/examples/sample](../../examples/sample/)

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_google"></a> [google](#requirement\_google) | 4.73.1 |
| <a name="requirement_google-beta"></a> [google-beta](#requirement\_google-beta) | 4.73.1 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | 4.73.1 |
| <a name="provider_google-beta"></a> [google-beta](#provider\_google-beta) | 4.73.1 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_artifact_registry"></a> [artifact\_registry](#input\_artifact\_registry) | Artifact Registry where the Docker images stored | <pre>object({<br>    project_id = string,<br>    location   = string,<br>    repository = string,<br>  })</pre> | n/a | yes |
| <a name="input_location"></a> [location](#input\_location) | Cloud Run Location of onboarding service | `string` | n/a | yes |
| <a name="input_prefix"></a> [prefix](#input\_prefix) | Resouce Prefix. If it's not empty, it should contains `-` as a last character eg. `dev-` | `string` | `""` | no |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | Google Cloud Project ID | `string` | n/a | yes |
| <a name="input_registry_encrypt_pub_key"></a> [registry\_encrypt\_pub\_key](#input\_registry\_encrypt\_pub\_key) | Encryption public key of the ONDC registry. This info should be avalable in the [ONDC onboarding document](https://github.com/ONDC-Official/developer-docs/blob/main/registry/Onboarding%20of%20Participants.md) | `string` | n/a | yes |
| <a name="input_request_id"></a> [request\_id](#input\_request\_id) | Request ID (ex. uuid). This should be the same ID you will use for `/subscribe` API. | `string` | n/a | yes |
| <a name="input_secret_id"></a> [secret\_id](#input\_secret\_id) | Secret Manager's Secret ID that store our key pairs | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_serverless_neg"></a> [serverless\_neg](#output\_serverless\_neg) | Serverless Network Endpoint Group |
| <a name="output_service"></a> [service](#output\_service) | Google Cloud Run Service |

## Resources

| Name | Type |
|------|------|
| [google-beta_google_compute_region_network_endpoint_group.serverless_neg](https://registry.terraform.io/providers/hashicorp/google-beta/4.73.1/docs/resources/google_compute_region_network_endpoint_group) | resource |
| [google_cloud_run_service.onboarding](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/cloud_run_service) | resource |
| [google_cloud_run_service_iam_member.invoker](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/cloud_run_service_iam_member) | resource |
| [google_project_service.cloud_run](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/project_service) | resource |
| [google_project_service.secret_manager](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/project_service) | resource |
| [google_secret_manager_secret_iam_member.read](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/secret_manager_secret_iam_member) | resource |
| [google_service_account.onboarding](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/service_account) | resource |

