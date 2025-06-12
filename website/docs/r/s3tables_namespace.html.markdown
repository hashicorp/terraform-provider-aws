---
subcategory: "S3 Tables"
layout: "aws"
page_title: "AWS: aws_s3tables_namespace"
description: |-
  Terraform resource for managing an Amazon S3 Tables Namespace.
---

# Resource: aws_s3tables_namespace

Terraform resource for managing an Amazon S3 Tables Namespace.

## Example Usage

### Basic Usage

```terraform
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

* `namespace` - (Required, Forces new resource) Name of the namespace.
  Must be between 1 and 255 characters in length.
  Can consist of lowercase letters, numbers, and underscores, and must begin and end with a lowercase letter or number.
* `table_bucket_arn` - (Required, Forces new resource) ARN referencing the Table Bucket that contains this Namespace.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Date and time when the namespace was created.
* `created_by` - Account ID of the account that created the namespace.
* `owner_account_id` - Account ID of the account that owns the namespace.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Tables Namespace using the `table_bucket_arn` and the value of `namespace`, separated by a semicolon (`;`). For example:

```terraform
import {
  to = aws_s3tables_namespace.example
  id = "arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket;example-namespace"
}
```

Using `terraform import`, import S3 Tables Namespace using the `table_bucket_arn` and the value of `namespace`, separated by a semicolon (`;`). For example:

```console
% terraform import aws_s3tables_namespace.example 'arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket;example-namespace'
```
