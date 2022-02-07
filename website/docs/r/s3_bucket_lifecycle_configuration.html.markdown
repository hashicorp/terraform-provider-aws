---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_lifecycle_configuration"
description: |-
  Provides a S3 bucket lifecycle configuration resource.
---

# Resource: aws_s3_bucket_lifecycle_configuration

Provides an independent configuration resource for S3 bucket [lifecycle configuration](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html).

## Example Usage

```terraform
resource "aws_s3_bucket" "bucket" {
  bucket = "my-bucket"
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "bucket-config" {
  bucket = aws_s3_bucket.bucket.bucket

  rule {
    id = "log"

    expiration {
      days = 90
    }

    filter {
      and {
        prefix = "log/"

        tags = {
          rule      = "log"
          autoclean = "true"
        }
      }
    }

    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "GLACIER"
    }
  }

  rule {
    id = "tmp"

    filter {
      prefix = "tmp/"
    }

    expiration {
      date = "2023-01-13T00:00:00Z"
    }

    status = "Enabled"
  }
}

resource "aws_s3_bucket" "versioning_bucket" {
  bucket = "my-versioning-bucket"
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "versioning" {
  bucket = aws_s3_bucket.versioning_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "versioning-bucket-config" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.versioning]

  bucket = aws_s3_bucket.versioning_bucket.bucket

  rule {
    id = "config"

    filter {
      prefix = "config/"
    }

    noncurrent_version_expiration {
      noncurrent_days = 90
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }

    noncurrent_version_transition {
      noncurrent_days = 60
      storage_class   = "GLACIER"
    }

    status = "Enabled"
  }
}
```

## Usage Notes

~> **NOTE:** To avoid conflicts always add the following lifecycle object to the `aws_s3_bucket` resource of the source bucket.

This resource implements the same features that are provided by the `lifecycle_rule` object of the [`aws_s3_bucket` resource](s3_bucket.html). To avoid conflicts or unexpected apply results, a lifecycle configuration is needed on the `aws_s3_bucket` to ignore changes to the internal `lifecycle_rule` object.  Failure to add the `lifecycle` configuration to the `aws_s3_bucket` will result in conflicting state results.

```
lifecycle {
  ignore_changes = [
    lifecycle_rule
  ]
}
```

The `aws_s3_bucket_lifecycle_configuration` resource provides the following features that are not available in the [`aws_s3_bucket` resource](s3_bucket.html):

* `filter` - Added to the `rule` configuration block [documented below](#filter).

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the source S3 bucket you want Amazon S3 to monitor.
* `expected_bucket_owner` - (Optional) The account ID of the expected bucket owner. If the bucket is owned by a different account, the request will fail with an HTTP 403 (Access Denied) error.
* `rule` - (Required) List of configuration blocks describing the rules managing the replication [documented below](#rule).

### rule

The `rule` configuration block supports the following arguments:

* `abort_incomplete_multipart_upload` - (Optional) Configuration block that specifies the days since the initiation of an incomplete multipart upload that Amazon S3 will wait before permanently removing all parts of the upload [documented below](#abort_incomplete_multipart_upload).
* `expiration` - (Optional) Configuration block that specifies the expiration for the lifecycle of the object in the form of date, days and, whether the object has a delete marker [documented below](#expiration).
* `filter` - (Optional, Required if `prefix` not specified) Configuration block used to identify objects that a Lifecycle Rule applies to [documented below](#filter).
* `id` - (Required) Unique identifier for the rule. The value cannot be longer than 255 characters.
* `noncurrent_version_expiration` - (Optional) Configuration block that specifies when noncurrent object versions expire [documented below](#noncurrent_version_expiration).
* `noncurrent_version_transition` - (Optional) Set of configuration blocks that specify the transition rule for the lifecycle rule that describes when noncurrent objects transition to a specific storage class [documented below](#noncurrent_version_transition).
* `prefix` - (Optional, Required if `filter` not specified) Prefix identifying one or more objects to which the rule applies. This has been deprecated by Amazon S3 and `filter` should be used instead.
* `status` - (Required) Whether the rule is currently being applied. Valid values: `Enabled` or `Disabled`.
* `transition` - (Optional) Set of configuration blocks that specify when an Amazon S3 object transitions to a specified storage class [documented below](#transition).

### abort_incomplete_multipart_upload

The `abort_incomplete_multipart_upload` configuration block supports the following arguments:

* `days_after_initiation` - The number of days after which Amazon S3 aborts an incomplete multipart upload.

### expiration

The `expiration` configuration block supports the following arguments:

* `date` - (Optional) The date the object is to be moved or deleted. Should be in GMT ISO 8601 Format.
* `days` - (Optional) The lifetime, in days, of the objects that are subject to the rule. The value must be a non-zero positive integer.
* `expired_object_delete_marker` - (Optional, Conflicts with `date` and `days`) Indicates whether Amazon S3 will remove a delete marker with no noncurrent versions. If set to `true`, the delete marker will be expired; if set to `false` the policy takes no action.

### filter

The `filter` configuration block supports the following arguments:

* `and`- (Optional) Configuration block used to apply a logical `AND` to two or more predicates. The Lifecycle Rule will apply to any object matching all of the predicates configured inside the `and` block.
* `object_size_greater_than` - (Optional) Minimum object size to which the rule applies.
* `object_size_less_than` - (Optional) Maximum object size to which the rule applies.
* `prefix` - (Optional) Prefix identifying one or more objects to which the rule applies.
* `tag` - (Optional) A configuration block for specifying a tag key and value [documented below](#tag).

### noncurrent_version_expiration

The `noncurrent_version_expiration` configuration block supports the following arguments:

* `newer_noncurrent_versions` - (Optional) The number of noncurrent versions Amazon S3 will retain. Must be a non-zero positive integer.
* `noncurrent_days` - (Optional) The number of days an object is noncurrent before Amazon S3 can perform the associated action. Must be a positive integer.

### noncurrent_version_transition

The `noncurrent_version_transition` configuration block supports the following arguments:

* `newer_noncurrent_versions` - (Optional) The number of noncurrent versions Amazon S3 will retain.
* `noncurrent_days` - (Optional) The number of days an object is noncurrent before Amazon S3 can perform the associated action.
* `storage_class` - (Required) The class of storage used to store the object. Valid Values: `GLACIER`, `STANDARD_IA`, `ONEZONE_IA`, `INTELLIGENT_TIERING`, `DEEP_ARCHIVE`, `GLACIER_IR`.

### transition

The `transition` configuration block supports the following arguments:

* `date` - (Optional) The date objects are transitioned to the specified storage class. The date value must be in ISO 8601 format. The time is always midnight UTC.
* `days` - (Optional) The number of days after creation when objects are transitioned to the specified storage class. The value must be a positive integer.
* `storage_class` - The class of storage used to store the object. Valid Values: `GLACIER`, `STANDARD_IA`, `ONEZONE_IA`, `INTELLIGENT_TIERING`, `DEEP_ARCHIVE`, `GLACIER_IR`.

### tag

The `tag` configuration block supports the following arguments:

* `key` - (Required) Name of the object key.
* `value` - (Required) Value of the tag.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`) if the latter is provided.

## Import

S3 bucket lifecycle configuration can be imported using the `bucket`, e.g.

```sh
$ terraform import aws_s3_bucket_lifecycle_configuration.example bucket-name
```

In addition, S3 bucket lifecycle configuration can be imported using the `bucket` and `expected_bucket_owner` separated by a comma (`,`) e.g.,

```
$ terraform import aws_s3_bucket_lifecycle_configuration.example bucket-name,123456789012
```
