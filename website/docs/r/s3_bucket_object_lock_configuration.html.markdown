---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_object_lock_configuration"
description: |-
  Provides an S3 bucket Object Lock configuration resource.
---

# Resource: aws_s3_bucket_object_lock_configuration

Provides an S3 bucket Object Lock configuration resource. For more information about Object Locking, go to [Using S3 Object Lock](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock.html) in the Amazon S3 User Guide.

-> This resource can be used enable Object Lock for **new** and **existing** buckets.

-> This resource cannot be used with S3 directory buckets.

## Example Usage

### Object Lock configuration for new or existing buckets

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "mybucket"
}

resource "aws_s3_bucket_versioning" "example" {
  bucket = aws_s3_bucket.example.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_object_lock_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    default_retention {
      mode = "COMPLIANCE"
      days = 5
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket` - (Required, Forces new resource) Name of the bucket.
* `expected_bucket_owner` - (Optional, Forces new resource) Account ID of the expected bucket owner.
* `object_lock_enabled` - (Optional, Forces new resource) Indicates whether this bucket has an Object Lock configuration enabled. Defaults to `Enabled`. Valid values: `Enabled`.
* `rule` - (Optional) Configuration block for specifying the Object Lock rule for the specified object. [See below](#rule).
* `token` - (Optional,Deprecated) This argument is deprecated and no longer needed to enable Object Lock.
To enable Object Lock for an existing bucket, you must first enable versioning on the bucket and then enable Object Lock. For more details on versioning, see the [`aws_s3_bucket_versioning` resource](s3_bucket_versioning.html.markdown).

### rule

The `rule` configuration block supports the following arguments:

* `default_retention` - (Required) Configuration block for specifying the default Object Lock retention settings for new objects placed in the specified bucket. [See below](#default_retention).

### default_retention

The `default_retention` configuration block supports the following arguments:

* `days` - (Optional, Required if `years` is not specified) Number of days that you want to specify for the default retention period.
* `mode` - (Required) Default Object Lock retention mode you want to apply to new objects placed in the specified bucket. Valid values: `COMPLIANCE`, `GOVERNANCE`.
* `years` - (Optional, Required if `days` is not specified) Number of years that you want to specify for the default retention period.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`) if the latter is provided.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 bucket Object Lock configuration using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```terraform
import {
  to = aws_s3_bucket_object_lock_configuration.example
  id = "bucket-name"
}
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

import {
  to = aws_s3_bucket_object_lock_configuration.example
  id = "bucket-name,123456789012"
}

**Using `terraform import` to import** S3 bucket Object Lock configuration using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```console
% terraform import aws_s3_bucket_object_lock_configuration.example bucket-name
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```console
% terraform import aws_s3_bucket_object_lock_configuration.example bucket-name,123456789012
```
