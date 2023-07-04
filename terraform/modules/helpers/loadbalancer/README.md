# Load Balancer Helper Module

Helper module to deploy Google Cloud's Application Load Balancer (Classic) for our services.

## Features
This module create a External Load Balancer and an URL maps for services.

## Example Usage
See the [terraform/examples/sample](../../../examples/sample/)

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_google"></a> [google](#requirement\_google) | 4.73.1 |
| <a name="requirement_google-beta"></a> [google-beta](#requirement\_google-beta) | 4.73.1 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | 4.73.1 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_address"></a> [address](#input\_address) | Load Balancer Address. If not specify, it will be created one | `string` | `""` | no |
| <a name="input_buyer"></a> [buyer](#input\_buyer) | Buyer | <pre>object({<br>    backend_name = optional(string)<br>    cluster_name = string<br>    network_name = string<br>    network_endpoint_groups = list(object({<br>      id = string<br>    }))<br>    max_rate_per_endpoint = number<br>  })</pre> | `null` | no |
| <a name="input_domains"></a> [domains](#input\_domains) | Managed SSL Certificate Domains | `list(string)` | `null` | no |
| <a name="input_env_prefix"></a> [env\_prefix](#input\_env\_prefix) | Environment, use as a prefix for each resource/service. | `string` | `""` | no |
| <a name="input_name"></a> [name](#input\_name) | Load Balancer Name | `string` | `"http-lb"` | no |
| <a name="input_onboarding"></a> [onboarding](#input\_onboarding) | Onboarding | <pre>object({<br>    backend_name = optional(string)<br>    serverless_neg = object({<br>      id = string<br>    })<br>    paths = optional(list(string))<br>  })</pre> | `null` | no |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | GCP Project ID | `string` | `""` | no |
| <a name="input_random_certificate_suffix"></a> [random\_certificate\_suffix](#input\_random\_certificate\_suffix) | Bool to enable/disable random certificate name generation. Set and keep this to true if you need to change the SSL cert. | `bool` | `false` | no |
| <a name="input_seller"></a> [seller](#input\_seller) | Seller | <pre>object({<br>    backend_name = optional(string)<br>    cluster_name = string<br>    network_name = string<br>    network_endpoint_groups = list(object({<br>      id = string<br>    }))<br>    max_rate_per_endpoint = number<br>  })</pre> | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_http_lb"></a> [http\_lb](#output\_http\_lb) | GCP Load Balancer Module |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_http_lb"></a> [http\_lb](#module\_http\_lb) | GoogleCloudPlatform/lb-http/google | ~> 9.0 |

## Resources

| Name | Type |
|------|------|
| [google_compute_backend_service.onboarding](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/compute_backend_service) | resource |
| [google_compute_url_map.default](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/compute_url_map) | resource |
| [google_client_config.main](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/data-sources/client_config) | data source |

