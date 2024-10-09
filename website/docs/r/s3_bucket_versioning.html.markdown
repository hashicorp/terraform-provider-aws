---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_versioning"
description: |-
  Provides an S3 bucket versioning resource.
---

# Resource: aws_s3_bucket_versioning

Provides a resource for controlling versioning on an S3 bucket.
Deleting this resource will either suspend versioning on the associated S3 bucket or
simply remove the resource from Terraform state if the associated S3 bucket is unversioned.

For more information, see [How S3 versioning works](https://docs.aws.amazon.com/AmazonS3/latest/userguide/manage-versioning-examples.html).

~> **NOTE:** If you are enabling versioning on the bucket for the first time, AWS recommends that you wait for 15 minutes after enabling versioning before issuing write operations (PUT or DELETE) on objects in the bucket.

-> This resource cannot be used with S3 directory buckets.

## Example Usage

### With Versioning Enabled

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example-bucket"
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "versioning_example" {
  bucket = aws_s3_bucket.example.id
  versioning_configuration {
    status = "Enabled"
  }
}
```

### With Versioning Disabled

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example-bucket"
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "versioning_example" {
  bucket = aws_s3_bucket.example.id
  versioning_configuration {
    status = "Disabled"
  }
}
```

### Object Dependency On Versioning

When you create an object whose `version_id` you need and an `aws_s3_bucket_versioning` resource in the same configuration, you are more likely to have success by ensuring the `s3_object` depends either implicitly (see below) or explicitly (i.e., using `depends_on = [aws_s3_bucket_versioning.example]`) on the `aws_s3_bucket_versioning` resource.

~> **NOTE:** For critical and/or production S3 objects, do not create a bucket, enable versioning, and create an object in the bucket within the same configuration. Doing so will not allow the AWS-recommended 15 minutes between enabling versioning and writing to the bucket.

This example shows the `aws_s3_object.example` depending implicitly on the versioning resource through the reference to `aws_s3_bucket_versioning.example.bucket` to define `bucket`:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yotto"
}

resource "aws_s3_bucket_versioning" "example" {
  bucket = aws_s3_bucket.example.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket_versioning.example.id
  key    = "droeloe"
  source = "example.txt"
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket` - (Required, Forces new resource) Name of the S3 bucket.
* `versioning_configuration` - (Required) Configuration block for the versioning parameters. [See below](#versioning_configuration).
* `expected_bucket_owner` - (Optional, Forces new resource) Account ID of the expected bucket owner.
* `mfa` - (Optional, Required if `versioning_configuration` `mfa_delete` is enabled) Concatenation of the authentication device's serial number, a space, and the value that is displayed on your authentication device.

### versioning_configuration

~> **Note:** While the `versioning_configuration.status` parameter supports `Disabled`, this value is only intended for _creating_ or _importing_ resources that correspond to unversioned S3 buckets.
Updating the value from `Enabled` or `Suspended` to `Disabled` will result in errors as the AWS S3 API does not support returning buckets to an unversioned state.

The `versioning_configuration` configuration block supports the following arguments:

* `status` - (Required) Versioning state of the bucket. Valid values: `Enabled`, `Suspended`, or `Disabled`. `Disabled` should only be used when creating or importing resources that correspond to unversioned S3 buckets.
* `mfa_delete` - (Optional) Specifies whether MFA delete is enabled in the bucket versioning configuration. Valid values: `Enabled` or `Disabled`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`) if the latter is provided.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 bucket versioning using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```terraform
import {
  to = aws_s3_bucket_versioning.example
  id = "bucket-name"
}
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```terraform
import {
  to = aws_s3_bucket_versioning.example
  id = "bucket-name,123456789012"
}
```

**Using `terraform import` to import** S3 bucket versioning using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```console
% terraform import aws_s3_bucket_versioning.example bucket-name
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```console
% terraform import aws_s3_bucket_versioning.example bucket-name,123456789012
```
