---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_object_lock_configuration"
description: |-
  Provides an S3 bucket Object Lock configuration resource.
---

# Resource: aws_s3_bucket_object_lock_configuration

Provides an S3 bucket Object Lock configuration resource. For more information about Object Locking, go to [Using S3 Object Lock](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock.html) in the Amazon S3 User Guide.

~> **NOTE:** This resource **does not enable** Object Lock for **new** buckets. It configures a default retention period for objects placed in the specified bucket.
Thus, to **enable** Object Lock for a **new** bucket, see the [Using object lock configuration](s3_bucket.html.markdown#Using-object-lock-configuration) section in  the `aws_s3_bucket` resource or the [Object Lock configuration for a new bucket](#object-lock-configuration-for-a-new-bucket) example below.
If you want to **enable** Object Lock for an **existing** bucket, contact AWS Support and see the [Object Lock configuration for an existing bucket](#object-lock-configuration-for-an-existing-bucket) example below.

## Example Usage

### Object Lock configuration for a new bucket

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "mybucket"

  object_lock_enabled = true
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

### Object Lock configuration for an existing bucket

This is a multistep process that requires AWS Support intervention.

1. Enable versioning on your S3 bucket, if you have not already done so.
Doing so will generate an "Object Lock token" in the back-end.

<!-- markdownlint-disable MD029 -->
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
```
<!-- markdownlint-disable MD029 -->

2. Contact AWS Support to provide you with the "Object Lock token" for the specified bucket and use the token (or token ID) within your new `aws_s3_bucket_object_lock_configuration` resource.
   Notice the `object_lock_enabled` argument does not need to be specified as it defaults to `Enabled`.

<!-- markdownlint-disable MD029 -->
```terraform
resource "aws_s3_bucket_object_lock_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    default_retention {
      mode = "COMPLIANCE"
      days = 5
    }
  }

  token = "NG2MKsfoLqV3A+aquXneSG4LOu/ekrlXkRXwIPFVfERT7XOPos+/k444d7RIH0E3W3p5QU6ml2exS2F/eYCFmMWHJ3hFZGk6al1sIJkmNhUMYmsv0jYVQyTTZNLM+DnfooA6SATt39mM1VW1yJh4E+XljMlWzaBwHKbss3/EjlGDjOmVhaSs4Z6427mMCaFD0RLwsYY7zX49gEc31YfOMJGxbXCXSeyNwAhhM/A8UH7gQf38RmjHjjAFbbbLtl8arsxTPW8F1IYohqwmKIr9DnotLLj8Tg44U2SPwujVaqmlKKP9s41rfgb4UbIm7khSafDBng0LGfxC4pMlT9Ny2w=="
}
```
<!-- markdownlint-disable MD029 -->

## Argument Reference

The following arguments are supported:

* `bucket` - (Required, Forces new resource) Name of the bucket.
* `expected_bucket_owner` - (Optional, Forces new resource) Account ID of the expected bucket owner.
* `object_lock_enabled` - (Optional, Forces new resource) Indicates whether this bucket has an Object Lock configuration enabled. Defaults to `Enabled`. Valid values: `Enabled`.
* `rule` - (Optional) Configuration block for specifying the Object Lock rule for the specified object. [See below](#rule).
* `token` - (Optional) Token to allow Object Lock to be enabled for an existing bucket. You must contact AWS support for the bucket's "Object Lock token".
The token is generated in the back-end when [versioning](https://docs.aws.amazon.com/AmazonS3/latest/userguide/manage-versioning-examples.html) is enabled on a bucket. For more details on versioning, see the [`aws_s3_bucket_versioning` resource](s3_bucket_versioning.html.markdown).

### rule

The `rule` configuration block supports the following arguments:

* `default_retention` - (Required) Configuration block for specifying the default Object Lock retention settings for new objects placed in the specified bucket. [See below](#default_retention).

### default_retention

The `default_retention` configuration block supports the following arguments:

* `days` - (Optional, Required if `years` is not specified) Number of days that you want to specify for the default retention period.
* `mode` - (Required) Default Object Lock retention mode you want to apply to new objects placed in the specified bucket. Valid values: `COMPLIANCE`, `GOVERNANCE`.
* `years` - (Optional, Required if `days` is not specified) Number of years that you want to specify for the default retention period.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`) if the latter is provided.

## Import

S3 bucket Object Lock configuration can be imported in one of two ways.

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider,
the S3 bucket Object Lock configuration resource should be imported using the `bucket` e.g.,

```
$ terraform import aws_s3_bucket_object_lock_configuration.example bucket-name
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider,
the S3 bucket Object Lock configuration resource should be imported using the `bucket` and `expected_bucket_owner` separated by a comma (`,`) e.g.,

```
$ terraform import aws_s3_bucket_object_lock_configuration.example bucket-name,123456789012
```
