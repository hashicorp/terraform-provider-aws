---
subcategory: "FinSpace"
layout: "aws"
page_title: "AWS: aws_finspace_kx_database"
description: |-
  Terraform resource for managing an AWS FinSpace Kx Database.
---

# Resource: aws_finspace_kx_database

Terraform resource for managing an AWS FinSpace Kx Database.

## Example Usage

### Basic Usage

```terraform
resource "aws_kms_key" "example" {
  description             = "Example KMS Key"
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "example" {
  name       = "my-tf-kx-environment"
  kms_key_id = aws_kms_key.example.arn
}

resource "aws_finspace_kx_database" "example" {
  environment_id = aws_finspace_kx_environment.example.id
  name           = "my-tf-kx-database"
  description    = "Example database description"
}
```

## Argument Reference

The following arguments are required:

* `environment_id` - (Required) Unique identifier for the KX environment.
* `name` - (Required) Name of the KX database.

The following arguments are optional:

* `description` - (Optional) Description of the KX database.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifier of the KX database.
* `created_timestamp` - Timestamp at which the databse is created in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `id` - A comma-delimited string joining environment ID and database name.
* `last_modified_timestamp` - Last timestamp at which the database was updated in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS FinSpace Kx Database using the `id` (environment ID and database name, comma-delimited). For example:

```terraform
import {
  to = aws_finspace_kx_database.example
  id = "n3ceo7wqxoxcti5tujqwzs,my-tf-kx-database"
}
```

Using `terraform import`, import an AWS FinSpace Kx Database using the `id` (environment ID and database name, comma-delimited). For example:

```console
% terraform import aws_finspace_kx_database.example n3ceo7wqxoxcti5tujqwzs,my-tf-kx-database
```
