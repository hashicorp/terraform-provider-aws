---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_accelerate_configuration"
description: |-
  Provides an S3 bucket accelerate configuration resource.
---

# Resource: aws_s3_bucket_accelerate_configuration

Provides an S3 bucket accelerate configuration resource. See the [Requirements for using Transfer Acceleration](https://docs.aws.amazon.com/AmazonS3/latest/userguide/transfer-acceleration.html#transfer-acceleration-requirements) for more details.

## Example Usage

```terraform
resource "aws_s3_bucket" "mybucket" {
  bucket = "mybucket"
}

resource "aws_s3_bucket_accelerate_configuration" "example" {
  bucket = aws_s3_bucket.mybucket.id
  status = "Enabled"
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket` - (Required, Forces new resource) Name of the bucket.
* `expected_bucket_owner` - (Optional, Forces new resource) Account ID of the expected bucket owner.
* `status` - (Required) Transfer acceleration state of the bucket. Valid values: `Enabled`, `Suspended`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`) if the latter is provided.

## Import

S3 bucket accelerate configuration can be imported using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```
$ terraform import aws_s3_bucket_accelerate_configuration.example bucket-name
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```
$ terraform import aws_s3_bucket_accelerate_configuration.example bucket-name,123456789012
```
