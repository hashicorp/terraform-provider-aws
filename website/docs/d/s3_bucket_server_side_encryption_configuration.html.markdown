---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_server_side_encryption_configuration"
description: |-
    Returns the default encryption configuration for an Amazon S3 bucket
---

# Data Source: aws_s3_bucket_server_side_encryption_configuration

Returns the default encryption configuration for an Amazon S3 bucket. By default, all buckets have a default encryption configuration that uses server-side encryption with Amazon S3 managed keys (SSE-S3).

## Example Usage

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "bucket.test.com"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

data "aws_s3_bucket_server_side_encryption_configuration" "example" {
  bucket = aws_s3_bucket.example.id
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) Name of the bucket

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Name of the bucket.
* `kms_master_key_id` - id of the KMS key used to encrypt bucket objects.
* `sse_algorithm` - `AES256` if default encryption is used, `aws:kms` if KMS is used.
* `bucket_key_enabled` - `true` if [Amazon S3 Bucket Keys](https://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-key.html) for SSE-KMS if used (only if `sse_algorithm` is `aws:kms`), `false` otherwise.
