---
subcategory: "S3 Tables"
layout: "aws"
page_title: "AWS: aws_s3tables_table"
description: |-
  Terraform resource for managing an Amazon S3 Tables Table.
---

# Resource: aws_s3tables_table

Terraform resource for managing an Amazon S3 Tables Table.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3tables_table" "example" {
  name             = "example_table"
  namespace        = aws_s3tables_namespace.example.namespace
  table_bucket_arn = aws_s3tables_namespace.example.table_bucket_arn
  format           = "ICEBERG"
}

resource "aws_s3tables_namespace" "example" {
  namespace        = "example_namespace"
  table_bucket_arn = aws_s3tables_table_bucket.example.arn
}

resource "aws_s3tables_table_bucket" "example" {
  name = "example-bucket"
}
```

## Argument Reference

The following arguments are required:

* `format` - (Required) Format of the table.
  Must be `ICEBERG`.
* `name` - (Required) Name of the table.
  Must be between 1 and 255 characters in length.
  Can consist of lowercase letters, numbers, and underscores, and must begin and end with a lowercase letter or number.
  A full list of table naming rules can be found in the [S3 Tables documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-tables-buckets-naming.html#naming-rules-table).
* `namespace` - (Required) Name of the namespace for this table.
  Must be between 1 and 255 characters in length.
  Can consist of lowercase letters, numbers, and underscores, and must begin and end with a lowercase letter or number.
* `table_bucket_arn` - (Required, Forces new resource) ARN referencing the Table Bucket that contains this Namespace.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `encryption_configuration` - (Optional) A single table bucket encryption configuration object.
  [See `encryption_configuration` below](#encryption_configuration).
* `maintenance_configuration` - (Optional) A single table bucket maintenance configuration object.
  [See `maintenance_configuration` below](#maintenance_configuration).

### `encryption_configuration`

The `encryption_configuration` object supports the following arguments:

* `kms_key_arn` - (Optional) The ARN of a KMS Key to be used with `aws:kms` `sse_algorithm`
* `sse_algorithm` - (Required) One of `aws:kms` or `AES256`

### `maintenance_configuration`

The `maintenance_configuration` object supports the following arguments:

* `iceberg_compaction` - (Required) A single Iceberg compaction settings object.
  [See `iceberg_compaction` below](#iceberg_compaction).
* `iceberg_snapshot_management` - (Required) A single Iceberg snapshot management settings object.
  [See `iceberg_snapshot_management` below](#iceberg_snapshot_management).

### `iceberg_compaction`

The `iceberg_compaction` object supports the following arguments:

* `settings` - (Required) Settings object for compaction.
  [See `iceberg_compaction.settings` below](#iceberg_compactionsettings).
* `status` - (Required) Whether the configuration is enabled.
  Valid values are `enabled` and `disabled`.

### `iceberg_compaction.settings`

The `iceberg_compaction.settings` object supports the following argument:

* `target_file_size_mb` - (Required) Data objects smaller than this size may be combined with others to improve query performance.
  Must be between `64` and `512`.

### `iceberg_snapshot_management`

The `iceberg_snapshot_management` configuration block supports the following arguments:

* `settings` - (Required) Settings object for snapshot management.
  [See `iceberg_snapshot_management.settings` below](#iceberg_snapshot_managementsettings).
* `status` - (Required) Whether the configuration is enabled.
  Valid values are `enabled` and `disabled`.

### `iceberg_snapshot_management.settings`

The `iceberg_snapshot_management.settings` object supports the following argument:

* `max_snapshot_age_hours` - (Required) Snapshots older than this will be marked for deletiion.
  Must be at least `1`.
* `min_snapshots_to_keep` - (Required) Minimum number of snapshots to keep.
  Must be at least `1`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the table.
* `created_at` - Date and time when the namespace was created.
* `created_by` - Account ID of the account that created the namespace.
* `metadata_location` - Location of table metadata.
* `modified_at` - Date and time when the namespace was last modified.
* `modified_by` - Account ID of the account that last modified the namespace.
* `owner_account_id` - Account ID of the account that owns the namespace.
* `type` - Type of the table.
  One of `customer` or `aws`.
* `version_token` - Identifier for the current version of table data.
* `warehouse_location` - S3 URI pointing to the S3 Bucket that contains the table data.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Tables Table using the `table_bucket_arn`, the value of `namespace`, and the value of `name`, separated by a semicolon (`;`). For example:

```terraform
import {
  to = aws_s3tables_table.example
  id = "arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket;example-namespace;example-table"
}
```

Using `terraform import`, import S3 Tables Table using the `table_bucket_arn`, the value of `namespace`, and the value of `name`, separated by a semicolon (`;`). For example:

```console
% terraform import aws_s3tables_table.example 'arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket;example-namespace;example-table'
```
