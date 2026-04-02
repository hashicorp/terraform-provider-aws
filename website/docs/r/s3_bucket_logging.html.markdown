---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_logging"
description: |-
  Provides an S3 bucket (server access) logging resource.
---

# Resource: aws_s3_bucket_logging

Provides an S3 bucket (server access) logging resource. For more information, see [Logging requests using server access logging](https://docs.aws.amazon.com/AmazonS3/latest/userguide/ServerLogs.html)
in the AWS S3 User Guide.

~> **Note:** Amazon S3 supports server access logging, AWS CloudTrail, or a combination of both. Refer to the [Logging options for Amazon S3](https://docs.aws.amazon.com/AmazonS3/latest/userguide/logging-with-S3.html)
to decide which method meets your requirements.

-> This resource cannot be used with S3 directory buckets.

## Example Usage

### Grant permission by using bucket policy

```terraform
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "logging" {
  bucket = "access-logging-bucket"
}

data "aws_iam_policy_document" "logging_bucket_policy" {
  statement {
    principals {
      identifiers = ["logging.s3.amazonaws.com"]
      type        = "Service"
    }
    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.logging.arn}/*"]
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}

resource "aws_s3_bucket_policy" "logging" {
  bucket = aws_s3_bucket.logging.bucket
  policy = data.aws_iam_policy_document.logging_bucket_policy.json
}

resource "aws_s3_bucket" "example" {
  bucket = "example-bucket"
}

resource "aws_s3_bucket_logging" "example" {
  bucket = aws_s3_bucket.example.bucket

  target_bucket = aws_s3_bucket.logging.bucket
  target_prefix = "log/"
  target_object_key_format {
    partitioned_prefix {
      partition_date_source = "EventTime"
    }
  }
}
```

### Grant permission by using bucket ACL

The [AWS Documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/enable-server-access-logging.html) does not recommend using the ACL.

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "my-tf-example-bucket"
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}

resource "aws_s3_bucket" "log_bucket" {
  bucket = "my-tf-log-bucket"
}

resource "aws_s3_bucket_acl" "log_bucket_acl" {
  bucket = aws_s3_bucket.log_bucket.id
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket_logging" "example" {
  bucket = aws_s3_bucket.example.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `bucket` - (Required, Forces new resource) Name of the bucket.
* `expected_bucket_owner` - (Optional, Forces new resource, **Deprecated**) Account ID of the expected bucket owner.
* `target_bucket` - (Required) Name of the bucket where you want Amazon S3 to store server access logs.
* `target_prefix` - (Required) Prefix for all log object keys.
* `target_grant` - (Optional) Set of configuration blocks with information for granting permissions. [See below](#target_grant).
* `target_object_key_format` - (Optional) Amazon S3 key format for log objects. [See below](#target_object_key_format).

### target_grant

The `target_grant` configuration block supports the following arguments:

* `grantee` - (Required) Configuration block for the person being granted permissions. [See below](#grantee).
* `permission` - (Required) Logging permissions assigned to the grantee for the bucket. Valid values: `FULL_CONTROL`, `READ`, `WRITE`.

### grantee

The `grantee` configuration block supports the following arguments:

* `email_address` - (Optional) Email address of the grantee. See [Regions and Endpoints](https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region) for supported AWS regions where this argument can be specified.
* `id` - (Optional) Canonical user ID of the grantee.
* `type` - (Required) Type of grantee. Valid values: `CanonicalUser`, `AmazonCustomerByEmail`, `Group`.
* `uri` - (Optional) URI of the grantee group.

### target_object_key_format

The `target_object_key_format` configuration block supports the following arguments:

* `partitioned_prefix` - (Optional) Partitioned S3 key for log objects, in the form `[target_prefix][SourceAccountId]/[SourceRegion]/[SourceBucket]/[YYYY]/[MM]/[DD]/[YYYY]-[MM]-[DD]-[hh]-[mm]-[ss]-[UniqueString]`. Conflicts with `simple_prefix`. [See below](#partitioned_prefix).
* `simple_prefix` - (Optional) Use the simple format for S3 keys for log objects, in the form `[target_prefix][YYYY]-[MM]-[DD]-[hh]-[mm]-[ss]-[UniqueString]`. To use, set `simple_prefix {}`. Conflicts with `partitioned_prefix`.

### partitioned_prefix

The `partitioned_prefix` configuration block supports the following arguments:

* `partition_date_source` - (Required) Specifies the partition date source for the partitioned prefix. Valid values: `EventTime`, `DeliveryTime`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`) if the latter is provided.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_s3_bucket_logging.example
  identity = {
    bucket = "bucket-name"
  }
}

resource "aws_s3_bucket_logging" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `bucket` (String) S3 bucket name.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 bucket logging using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```terraform
import {
  to = aws_s3_bucket_logging.example
  id = "bucket-name"
}
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```terraform
import {
  to = aws_s3_bucket_logging.example
  id = "bucket-name,123456789012"
}
```

**Using `terraform import` to import** S3 bucket logging using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```console
% terraform import aws_s3_bucket_logging.example bucket-name
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```console
% terraform import aws_s3_bucket_logging.example bucket-name,123456789012
```
