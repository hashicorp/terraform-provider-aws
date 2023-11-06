---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_named_query"
description: |-
  Provides an Athena Named Query resource.
---

# Resource: aws_athena_named_query

Provides an Athena Named Query resource.

## Example Usage

```terraform
resource "aws_s3_bucket" "hoge" {
  bucket = "tf-test"
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = "Athena KMS Key"
}

resource "aws_athena_workgroup" "test" {
  name = "example"

  configuration {
    result_configuration {
      encryption_configuration {
        encryption_option = "SSE_KMS"
        kms_key_arn       = aws_kms_key.test.arn
      }
    }
  }
}

resource "aws_athena_database" "hoge" {
  name   = "users"
  bucket = aws_s3_bucket.hoge.id
}

resource "aws_athena_named_query" "foo" {
  name      = "bar"
  workgroup = aws_athena_workgroup.test.id
  database  = aws_athena_database.hoge.name
  query     = "SELECT * FROM ${aws_athena_database.hoge.name} limit 10;"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Plain language name for the query. Maximum length of 128.
* `workgroup` - (Optional) Workgroup to which the query belongs. Defaults to `primary`
* `database` - (Required) Database to which the query belongs.
* `query` - (Required) Text of the query itself. In other words, all query statements. Maximum length of 262144.
* `description` - (Optional) Brief explanation of the query. Maximum length of 1024.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique ID of the query.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Athena Named Query using the query ID. For example:

```terraform
import {
  to = aws_athena_named_query.example
  id = "0123456789"
}
```

Using `terraform import`, import Athena Named Query using the query ID. For example:

```console
% terraform import aws_athena_named_query.example 0123456789
```
