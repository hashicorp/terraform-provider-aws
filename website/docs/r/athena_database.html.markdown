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
  bucket = aws_s3_bucket.example.bucket
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the database to create.
* `bucket` - (Required) Name of s3 bucket to save the results of the query execution.
* `expected_bucket_owner` - (Optional) The AWS account ID that you expect to be the owner of the Amazon S3 bucket.
* `encryption_configuration` - (Optional) The encryption key block AWS Athena uses to decrypt the data in S3, such as an AWS Key Management Service (AWS KMS) key. See [Encryption Configuration](#encryption-configuration) Below.
* `acl_configuration` - (Optional) Indicates that an Amazon S3 canned ACL should be set to control ownership of stored query results. See [ACL Configuration](#acl-configuration) Below.
* `force_destroy` - (Optional, Default: false) A boolean that indicates all tables should be deleted from the database so that the database can be destroyed without error. The tables are *not* recoverable.

### Encryption Configuration

* `encryption_option` - (Required) The type of key; one of `SSE_S3`, `SSE_KMS`, `CSE_KMS`
* `kms_key` - (Optional) The KMS key ARN or ID; required for key types `SSE_KMS` and `CSE_KMS`.

### ACL Configuration

* `s3_acl_option` - (Required) The Amazon S3 canned ACL that Athena should specify when storing query results. Valid value is `BUCKET_OWNER_FULL_CONTROL`.

~> **NOTE:** When Athena queries are executed, result files may be created in the specified bucket. Consider using `force_destroy` on the bucket too in order to avoid any problems when destroying the bucket.  

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The database name
