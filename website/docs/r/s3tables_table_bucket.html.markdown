---
subcategory: "S3 Tables"
layout: "aws"
page_title: "AWS: aws_s3tables_table_bucket"
description: |-
  Terraform resource for managing an Amazon S3 Tables Table Bucket.
---

# Resource: aws_s3tables_table_bucket

Terraform resource for managing an Amazon S3 Tables Table Bucket.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3tables_table_bucket" "example" {
  name = "example-bucket"
}
```

## Argument Reference

The following argument is required:

* `name` - (Required, Forces new resource) Name of the table bucket.
  Must be between 3 and 63 characters in length.
  Can consist of lowercase letters, numbers, and hyphens, and must begin and end with a lowercase letter or number.
  A full list of bucket naming rules may be found in [S3 Tables documentation](???).

The following argument is optional:

* `maintenance_configuration` - (Optional) A single table bucket maintenance configuration block.
  [See `maintenance_configuration` below](#maintenance_configuration)

### maintenance_configuration

The `maintenance_configuration` configuration block supports the following argument:

* `iceberg_unreferenced_file_removal` - (Required) A single Iceberg unreferenced file removal settings block.
  [See `iceberg_unreferenced_file_removal` below](#iceberg_unreferenced_file_removal)

### `iceberg_unreferenced_file_removal`

The `iceberg_unreferenced_file_removal` configuration block supports the following arguments:

* `settings` - (Required) Settings for unreferenced file removal.
  [See `iceberg_unreferenced_file_removal.settings` below](#iceberg_unreferenced_file_removalsettings)
* `status` - (Required) Whether the configuration is enabled.
  Valid values are `enabled` and `disabled`.

### `iceberg_unreferenced_file_removal.settings`

The `iceberg_unreferenced_file_removal.settings` configuration block supports the following arguments:

* `non_current_days` - (Required) Data objects marked for deletion are deleted after this many days.
  Must be at least `1`.
* `unreferenced_days` - (Required) Unreferenced data objects are marked for deletion after this many days.
  Must be at least `1`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the table bucket.
* `created_at` - Date and time when the bucket was created.
* `owner_account_id` - Account ID of the account that owns the table bucket.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Tables Table Bucket using the `arn`. For example:

```terraform
import {
  to = aws_s3tables_table_bucket.example
  id = "arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket"
}
```

Using `terraform import`, import S3 Tables Table Bucket using the `arn`. For example:

```console
% terraform import aws_s3tables_table_bucket.example arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket
```
