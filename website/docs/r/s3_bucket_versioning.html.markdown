---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_versioning"
description: |-
  Provides an S3 bucket versioning resource.
---

# Resource: aws_s3_bucket_versioning

Provides a resource for controlling versioning on an S3 bucket.
Deleting this resource will suspend versioning on the associated S3 bucket.
For more information, see [How S3 versioning works](https://docs.aws.amazon.com/AmazonS3/latest/userguide/manage-versioning-examples.html).

~> **NOTE:** If you are enabling versioning on the bucket for the first time, AWS recommends that you wait for 15 minutes after enabling versioning before issuing write operations (PUT or DELETE) on objects in the bucket.

## Example Usage

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example-bucket"
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "versioning_example" {
  bucket = aws_s3_bucket.example.id
  versioning_configuration {
    status = "Enabled"
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required, Forces new resource) The name of the S3 bucket.
* `versioning_configuration` - (Required) Configuration block for the versioning parameters [detailed below](#versioning_configuration).
* `expected_bucket_owner` - (Optional, Forces new resource) The account ID of the expected bucket owner.
* `mfa` - (Optional, Required if `versioning_configuration` `mfa_delete` is enabled) The concatenation of the authentication device's serial number, a space, and the value that is displayed on your authentication device.

### versioning_configuration

The `versioning_configuration` configuration block supports the following arguments:

* `status` - (Required) The versioning state of the bucket. Valid values: `Enabled` or `Suspended`.
* `mfa_delete` - (Optional) Specifies whether MFA delete is enabled in the bucket versioning configuration. Valid values: `Enabled` or `Disabled`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`) if the latter is provided.

## Import

S3 bucket versioning can be imported using the `bucket`, e.g.

```
$ terraform import aws_s3_bucket_versioning.example bucket-name
```

In addition, S3 bucket versioning can be imported using the `bucket` and `expected_bucket_owner` separated by a comma (`,`), e.g.

```
$ terraform import aws_s3_bucket_versioning.example bucket-name,123456789012
```
