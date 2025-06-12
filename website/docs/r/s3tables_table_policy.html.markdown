---
subcategory: "S3 Tables"
layout: "aws"
page_title: "AWS: aws_s3tables_table_policy"
description: |-
  Terraform resource for managing an Amazon S3 Tables Table Policy.
---

# Resource: aws_s3tables_table_policy

Terraform resource for managing an Amazon S3 Tables Table Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3tables_table_policy" "example" {
  resource_policy  = data.aws_iam_policy_document.example.json
  name             = aws_s3tables_table.test.name
  namespace        = aws_s3tables_table.test.namespace
  table_bucket_arn = aws_s3tables_table.test.table_bucket_arn
}

data "aws_iam_policy_document" "example" {
  statement {
    # ...
  }
}

resource "aws_s3tables_table" "example" {
  name             = "example_table"
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

* `resource_policy` - (Required) Amazon Web Services resource-based policy document in JSON format.
* `name` - (Required, Forces new resource) Name of the table.
  Must be between 1 and 255 characters in length.
  Can consist of lowercase letters, numbers, and underscores, and must begin and end with a lowercase letter or number.
* `namespace` - (Required, Forces new resource) Name of the namespace for this table.
  Must be between 1 and 255 characters in length.
  Can consist of lowercase letters, numbers, and underscores, and must begin and end with a lowercase letter or number.
* `table_bucket_arn` - (Required, Forces new resource) ARN referencing the Table Bucket that contains this Namespace.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Tables Table Policy using the `table_bucket_arn`, the value of `namespace`, and the value of `name`, separated by a semicolon (`;`). For example:

```terraform
import {
  to = aws_s3tables_table_policy.example
  id = "arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket;example-namespace;example-table"
}
```

Using `terraform import`, import S3 Tables Table Policy using the `table_bucket_arn`, the value of `namespace`, and the value of `name`, separated by a semicolon (`;`). For example:

```console
% terraform import aws_s3tables_table_policy.example 'arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket;example-namespace;example-table'
```
