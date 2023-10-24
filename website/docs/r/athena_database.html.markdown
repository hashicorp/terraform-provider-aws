---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_database"
description: |-
  Provides an Athena database.
---

# Resource: aws_athena_database

Provides an Athena database.

## Example Usage

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_athena_database" "example" {
  name   = "database_name"
  bucket = aws_s3_bucket.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket` - (Required) Name of S3 bucket to save the results of the query execution.
* `name` - (Required) Name of the database to create.
* `acl_configuration` - (Optional) That an Amazon S3 canned ACL should be set to control ownership of stored query results. See [ACL Configuration](#acl-configuration) below.
* `comment` - (Optional) Description of the database.
* `encryption_configuration` - (Optional) Encryption key block AWS Athena uses to decrypt the data in S3, such as an AWS Key Management Service (AWS KMS) key. See [Encryption Configuration](#encryption-configuration) below.
* `expected_bucket_owner` - (Optional) AWS account ID that you expect to be the owner of the Amazon S3 bucket.
* `force_destroy` - (Optional, Default: false) Boolean that indicates all tables should be deleted from the database so that the database can be destroyed without error. The tables are *not* recoverable.
* `properties` - (Optional) Key-value map of custom metadata properties for the database definition.

### ACL Configuration

* `s3_acl_option` - (Required) Amazon S3 canned ACL that Athena should specify when storing query results. Valid value is `BUCKET_OWNER_FULL_CONTROL`.

~> **NOTE:** When Athena queries are executed, result files may be created in the specified bucket. Consider using `force_destroy` on the bucket too in order to avoid any problems when destroying the bucket.  

### Encryption Configuration

* `encryption_option` - (Required) Type of key; one of `SSE_S3`, `SSE_KMS`, `CSE_KMS`
* `kms_key` - (Optional) KMS key ARN or ID; required for key types `SSE_KMS` and `CSE_KMS`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Database name

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Athena Databases using their name. For example:

```terraform
import {
  to = aws_athena_database.example
  id = "example"
}
```

Using `terraform import`, import Athena Databases using their name. For example:

```console
% terraform import aws_athena_database.example example
```

Certain resource arguments, like `encryption_configuration` and `bucket`, do not have an API method for reading the information after creation. If the argument is set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to hide the difference. For example:

```terraform
resource "aws_athena_database" "example" {
  name   = "database_name"
  bucket = aws_s3_bucket.example.id

  # There is no API for reading bucket
  lifecycle {
    ignore_changes = [bucket]
  }
}
```
