---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_source_api_association"
description: |-
  Terraform resource for managing an AWS AppSync Source Api Association.
---
# Resource: aws_appsync_source_api_association

Terraform resource for managing an AWS AppSync Source Api Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_appsync_source_api_association" "test" {
  description   = "My source API Merged"
  merged_api_id = "gzos6bteufdunffzzifiowisoe"
  source_api_id = "fzzifiowisoegzos6bteufdunf"
}
```

## Argument Reference

The following arguments are optional:

* `description` - (Optional) Description of the source API being merged.
* `merged_api_arn` - (Optional) ARN of the merged API. One of `merged_api_arn` or `merged_api_id` must be specified.
* `merged_api_id` - (Optional) ID of the merged API. One of `merged_api_arn` or `merged_api_id` must be specified.
* `source_api_arn` - (Optional) ARN of the source API. One of `source_api_arn` or `source_api_id` must be specified.
* `source_api_id` - (Optional) ID of the source API. One of `source_api_arn` or `source_api_id` must be specified.

### `source_api_association_config` Block

The `source_api_association_config` configuration block supports the following arguments:

* `merge_type` - (Required) Merge type. Valid values: `MANUAL_MERGE`, `AUTO_MERGE`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Source Api Association.
* `association_id` - ID of the Source Api Association.
* `id` - Combined ID of the Source Api Association and Merge Api.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppSync Source Api Association using the `example_id_arg`. For example:

```terraform
import {
  to = aws_appsync_source_api_association.example
  id = "gzos6bteufdunffzzifiowisoe,243685a0-9347-4a1a-89c1-9b57dea01e31"
}
```

Using `terraform import`, import AppSync Source Api Association using the `gzos6bteufdunffzzifiowisoe,243685a0-9347-4a1a-89c1-9b57dea01e31`. For example:

```console
% terraform import aws_appsync_source_api_association.example gzos6bteufdunffzzifiowisoe,243685a0-9347-4a1a-89c1-9b57dea01e31
```
