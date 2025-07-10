---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_ownership_controls"
description: |-
  Manages S3 Bucket Ownership Controls.
---

# Resource: aws_s3_bucket_ownership_controls

Provides a resource to manage S3 Bucket Ownership Controls. For more information, see the [S3 Developer Guide](https://docs.aws.amazon.com/AmazonS3/latest/dev/about-object-ownership.html).

-> This resource cannot be used with S3 directory buckets.

## Example Usage

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_ownership_controls" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `bucket` - (Required) Name of the bucket that you want to associate this access point with.
* `rule` - (Required) Configuration block(s) with Ownership Controls rules. Detailed below.

### rule Configuration Block

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `object_ownership` - (Required) Object ownership. Valid values: `BucketOwnerPreferred`, `ObjectWriter` or `BucketOwnerEnforced`
    * `BucketOwnerPreferred` - Objects uploaded to the bucket change ownership to the bucket owner if the objects are uploaded with the `bucket-owner-full-control` canned ACL.
    * `ObjectWriter` - Uploading account will own the object if the object is uploaded with the `bucket-owner-full-control` canned ACL.
    * `BucketOwnerEnforced` - Bucket owner automatically owns and has full control over every object in the bucket. ACLs no longer affect permissions to data in the S3 bucket.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - S3 Bucket name.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Bucket Ownership Controls using S3 Bucket name. For example:

```terraform
import {
  to = aws_s3_bucket_ownership_controls.example
  id = "my-bucket"
}
```

Using `terraform import`, import S3 Bucket Ownership Controls using S3 Bucket name. For example:

```console
% terraform import aws_s3_bucket_ownership_controls.example my-bucket
```
