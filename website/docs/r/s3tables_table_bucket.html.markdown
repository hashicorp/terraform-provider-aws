---
subcategory: "S3 Tables"
layout: "aws"
page_title: "AWS: aws_s3tables_table_bucket"
description: |-
  Terraform resource for managing an AWS S3 Tables Table Bucket.
---

# Resource: aws_s3tables_table_bucket

Terraform resource for managing an AWS S3 Tables Table Bucket.

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

## Attribute Reference

This resource exports the following attributes in addition to the argument above:

* `arn` - ARN of the table bucket.
* `created_at` - Date and time when the bucket was created.
* `owner_account_id` - Account ID of the account that owns the table bucket.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Tables Table Bucket using the `arn`.
For example:

```terraform
import {
  to = aws_s3tables_table_bucket.example
  id = "arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket"
}
```

Using `terraform import`, import S3 Tables Table Bucket using the `arn`.
For example:

```console
% terraform import aws_s3tables_table_bucket.example arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket
```
