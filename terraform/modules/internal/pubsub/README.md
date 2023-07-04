# Pub/Sub Module (Internal)

Terraform module to create the Pub/Sub topics and subscriptions for ONDC Accelerator services.

## Example Usage
```hcl
module "sample_key_rotation" {
  source = "../../modules/key-rotation"

  artifact_registry = {
    project_id = "example-project"
    location   = "us-central1"
    repository = "example-repo"
  }

  prefix    = "sample-keys"
  secret_id = "sample-key-rotation-keys"
  service_accounts = [
    google_service_account.sample_buyer_cluster.email,
    google_service_account.sample_seller_cluster.email,
  ]

  registry_url = "https://registry.example.com"
}
```
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_google"></a> [google](#requirement\_google) | 4.73.1 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | 4.73.1 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [google_project_service.pubsub](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/project_service) | resource |
| [google_pubsub_subscription.callback](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/pubsub_subscription) | resource |
| [google_pubsub_subscription.send](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/pubsub_subscription) | resource |
| [google_pubsub_topic.callback](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/pubsub_topic) | resource |
| [google_pubsub_topic.send](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/pubsub_topic) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_callback_message_retention_duration"></a> [callback\_message\_retention\_duration](#input\_callback\_message\_retention\_duration) | callback message retention duration. | `string` | `"600s"` | no |
| <a name="input_prefix"></a> [prefix](#input\_prefix) | topics & subscriptions prefix. | `string` | `"google-open-commerce"` | no |
| <a name="input_send_message_retention_duration"></a> [send\_message\_retention\_duration](#input\_send\_message\_retention\_duration) | send message retention duration. | `string` | `"600s"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_prefix"></a> [prefix](#output\_prefix) | Pub/Sub resource prefix |
| <a name="output_topic"></a> [topic](#output\_topic) | Pub/Sub topic |
<!-- END_TF_DOCS -->