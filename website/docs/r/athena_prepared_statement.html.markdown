---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_prepared_statement"
description: |-
  Terraform resource for managing an AWS Athena Prepared Statement.
---

# Resource: aws_athena_prepared_statement

Terraform resource for managing an Athena Prepared Statement.

## Example Usage

```terraform
resource "aws_s3_bucket" "test" {
  bucket        = "tf-test"
  force_destroy = true
}

resource "aws_athena_workgroup" "test" {
  name = "tf-test"
}

resource "aws_athena_database" "test" {
  name   = "example"
  bucket = aws_s3_bucket.test.bucket
}

resource "aws_athena_prepared_statement" "test" {
  name            = "tf_test"
  query_statement = "SELECT * FROM ${aws_athena_database.test.name} WHERE x = ?"
  workgroup       = aws_athena_workgroup.test.name
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the prepared statement. Maximum length of 256.
* `workgroup` - (Required) The name of the workgroup to which the prepared statement belongs.
* `query_statement` - (Required) The query string for the prepared statement.
* `description` - (Optional) Brief explanation of prepared statement. Maximum length of 1024.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the prepared statement

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Athena Prepared Statement using the `WORKGROUP-NAME/STATEMENT-NAME`. For example:

```terraform
import {
  to = aws_athena_prepared_statement.example
  id = "12345abcde/example"
}
```

Using `terraform import`, import Athena Prepared Statement using the `WORKGROUP-NAME/STATEMENT-NAME`. For example:

```console
% terraform import aws_athena_prepared_statement.example 12345abcde/example 
```
