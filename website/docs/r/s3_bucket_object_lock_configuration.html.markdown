---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_object_lock_configuration"
description: |-
  Provides an S3 bucket Object Lock configuration resource.
---

# Resource: aws_s3_bucket_object_lock_configuration

Provides an S3 bucket Object Lock configuration resource. For more information about Object Locking, go to [Using S3 Object Lock](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock.html) in the Amazon S3 User Guide.

~> **NOTE:** You can only enable Object Lock for new buckets. If you want to turn on Object Lock for an existing bucket, contact AWS Support.

## Example Usage

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "mybucket"

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}

resource "aws_s3_bucket_object_lock_configuration" "example" {
  bucket = aws_s3_bucket.example.bucket

  rule {
    default_retention {
      mode = "COMPLIANCE"
      days = 5
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required, Forces new resource) The name of the bucket.
* `expected_bucket_owner` - (Optional, Forces new resource) The account ID of the expected bucket owner.
* `object_lock_enabled` - (Optional, Forces new resource) Indicates whether this bucket has an Object Lock configuration enabled. Defaults to `Enabled`. Valid values: `Enabled`.
* `rule` - (Required) Configuration block for specifying the Object Lock rule for the specified object [detailed below](#rule).
* `token` - (Optional) A token to allow Object Lock to be enabled for an existing bucket.

### rule

The `rule` configuration block supports the following arguments:

* `default_retention` - (Required) A configuration block for specifying the default Object Lock retention settings for new objects placed in the specified bucket [detailed below](#default_retention).

### default_retention

The `default_retention` configuration block supports the following arguments:

* `days` - (Optional, Required if `years` is not specified) The number of days that you want to specify for the default retention period.
* `mode` - (Required) The default Object Lock retention mode you want to apply to new objects placed in the specified bucket. Valid values: `COMPLIANCE`, `GOVERNANCE`.
* `years` - (Optional, Required if `days` is not specified) The number of years that you want to specify for the default retention period.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`) if the latter is provided.

## Import

S3 bucket Object Lock configuration can be imported using the `bucket` e.g.,

```
$ terraform import aws_s3_bucket_object_lock_configuration.example bucket-name
```

In addition, S3 bucket Object Lock configuration can be imported using the `bucket` and `expected_bucket_owner` separated by a comma (`,`) e.g.,

```
$ terraform import aws_s3_bucket_object_lock_configuration.example bucket-name,123456789012
```
