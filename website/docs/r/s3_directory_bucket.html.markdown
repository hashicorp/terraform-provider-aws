---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_directory_bucket"
description: |-
  Provides an Amazon S3 Express directory bucket resource.
---

# Resource: aws_s3_directory_bucket

Provides an Amazon S3 Express directory bucket resource.

## Example Usage

```terraform
resource "aws_s3_directory_bucket" "example" {
  bucket = "example--usw2-az1--x-s3"

  location {
    name = "usw2-az1"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket` - (Required) Name of the bucket. The name must be in the format `[bucket_name]--[azid]--x-s3`. Use the [`aws_s3_bucket`](s3_bucket.html) resource to manage general purpose buckets.
* `data_redundancy` - (Optional, Default:`SingleAvailabilityZone`) Data redundancy. Valid values: `SingleAvailabilityZone`.
* `force_destroy` - (Optional, Default:`false`) Boolean that indicates all objects should be deleted from the bucket *when the bucket is destroyed* so that the bucket can be destroyed without error. These objects are *not* recoverable. This only deletes objects when the bucket is destroyed, *not* when setting this parameter to `true`. Once this parameter is set to `true`, there must be a successful `terraform apply` run before a destroy is required to update this value in the resource state. Without a successful `terraform apply` after this parameter is set, this flag will have no effect. If setting this field in the same operation that would require replacing the bucket or destroying the bucket, this flag will not work. Additionally when importing a bucket, a successful `terraform apply` is required to set this value in state before it will take effect on a destroy operation.
* `location` - (Required) Bucket location. See [Location](#location) below for more details.
* `type` - (Optional, Default:`Directory`) Bucket type. Valid values: `Directory`.

### Location

The `location` block supports the following:

* `name` - (Required) [Availability Zone ID](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#az-ids) or Local Zone ID.
* `type` - (Optional, Default:`AvailabilityZone`) Location type. Valid values: `AvailabilityZone`, `LocalZone`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - (**Deprecated**, use `bucket` instead) Name of the bucket.
* `arn` - ARN of the bucket.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Amazon S3 Express directory bucket using `bucket`. For example:

```terraform
import {
  to = aws_s3_directory_bucket.example
  id = "example--usw2-az1--x-s3"
}
```

Using `terraform import`, import S3 bucket using `bucket`. For example:

```console
% terraform import aws_s3_directory_bucket.example example--usw2-az1--x-s3
```
