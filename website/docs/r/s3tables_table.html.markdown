---
subcategory: "S3 Tables"
layout: "aws"
page_title: "AWS: aws_s3tables_table"
description: |-
  Terraform resource for managing an AWS S3 Tables Table.
---

# Resource: aws_s3tables_table

Terraform resource for managing an AWS S3 Tables Table.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3tables_table" "example" {
  name             = "example-table"
  namespace        = aws_s3tables_namespace.example
  table_bucket_arn = aws_s3tables_namespace.example.table_bucket_arn
  format           = "ICEBERG"
}

resource "aws_s3tables_namespace" "example" {
  namespace        = ["example-namespace"]
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
* `namespace` - (Required) Name of the namespace for this table.
  Must be between 1 and 255 characters in length.
  Can consist of lowercase letters, numbers, and underscores, and must begin and end with a lowercase letter or number.
* `table_bucket_arn` - (Required, Forces new resource) ARN referencing the Table Bucket that contains this Namespace.

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

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Tables Table using the `table_bucket_arn`, the value of `namespace`, and the value of `name`, separated by a semicolon (`;`).
For example:

```terraform
import {
  to = aws_s3tables_table.example
  id = "arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket;example-namespace;example-table"
}
```

Using `terraform import`, import S3 Tables Table using the `example_id_arg`.
For example:

```console
% terraform import aws_s3tables_table.example 'arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket;example-namespace;example-table'
```
