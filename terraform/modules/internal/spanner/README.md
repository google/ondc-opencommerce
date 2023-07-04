# Spanner Module (Internal)

Terraform module to create a spanner database for ONDC Accelerator services. 

## Database Schema

1. Transaction table

| Column          | Data Type                       | Description                                                                                                                            |
|-----------------|---------------------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| TransactionID   | STRING(36) NOT NULL PRIMARY KEY |                                                                                                                                        |
| TransactionType | INT64 NOT NULL PRIMARY KEY      | Maps to a transaction type (1 - REQUEST-ACTION, 2 - CALLBACK-ACTION)                                                                   |
| TransactionAPI  | INT64 NOT NULL PRIMARY KEY      | Maps to a message type (eg. 1 - search, 2 - on_search)                                                                                 |
| MessageID       | STRING(36) NOT NULL PRIMARY KEY |                                                                                                                                        |
| RequestID       | STRING(36)                      | Generated UUID to make the row unique.                                                                                                 |
| Payload         | JSON NOT NULL                   | The Payload of the transaction                                                                                                         |
| ProviderID      | STRING(255)                     | Subscriber ID of the Entity who provide the payload                                                                                    |
| MessageStatus   | STRING(255)                     | ACK, NACK                                                                                                                              |
| ErrorType       | STRING(36)                      |                                                                                                                                        |
| ErrorCode       | STRING(255)                     | An error code to indicate the cause of an error                                                                                        |
| ErrorPath       | STRING(MAX)                     |                                                                                                                                        |
| ErrorMessage    | STRING(MAX)                     | A longer and descriptive error messsage providing full details of an error                                                             |
| ReqReceivedTime | TIMESTAMP                       | The timestamp at which a request was received                                                                                          |
| AdditionalData  | JSON                            | Any additional data about the transaction that is primarily used to show information about the transaction (and not used for querying) |


API Mapping

| API        | Value |
|------------|-------|
| search     | 1     |
| on_search  | 2     |
| select     | 3     |
| on_select  | 4     |
| init       | 5     |
| on_init    | 6     |
| confirm    | 7     |
| on_confirm | 8     |
| status     | 9     |
| on_status  | 10    |
| track      | 11    |
| on_track   | 12    |
| cancel     | 13    |
| on_cancel  | 14    |
| update     | 15    |
| on_update  | 16    |
| rating     | 17    |
| on_rating  | 18    |
| support    | 19    |
| on_support | 20    |

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
| [google_project_service.spanner](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/project_service) | resource |
| [google_spanner_database.spanner_database](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/spanner_database) | resource |
| [google_spanner_instance.spanner_instance](https://registry.terraform.io/providers/hashicorp/google/4.73.1/docs/resources/spanner_instance) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_database_name"></a> [database\_name](#input\_database\_name) | n/a | `string` | `""` | no |
| <a name="input_display_name"></a> [display\_name](#input\_display\_name) | n/a | `string` | `""` | no |
| <a name="input_force_destroy"></a> [force\_destroy](#input\_force\_destroy) | n/a | `bool` | `false` | no |
| <a name="input_instance_config"></a> [instance\_config](#input\_instance\_config) | n/a | `string` | `""` | no |
| <a name="input_instance_name"></a> [instance\_name](#input\_instance\_name) | n/a | `string` | `""` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_database"></a> [database](#output\_database) | n/a |
| <a name="output_instance"></a> [instance](#output\_instance) | n/a |
<!-- END_TF_DOCS -->
