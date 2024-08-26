---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_metric"
description: |-
  Provides a S3 bucket metrics configuration resource.
---

# Resource: aws_s3_bucket_metric

Provides a S3 bucket [metrics configuration](http://docs.aws.amazon.com/AmazonS3/latest/dev/metrics-configurations.html) resource.

-> This resource cannot be used with S3 directory buckets.

## Example Usage

### Add metrics configuration for entire S3 bucket

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_metric" "example-entire-bucket" {
  bucket = aws_s3_bucket.example.id
  name   = "EntireBucket"
}
```

### Add metrics configuration with S3 object filter

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_metric" "example-filtered" {
  bucket = aws_s3_bucket.example.id
  name   = "ImportantBlueDocuments"

  filter {
    prefix = "documents/"

    tags = {
      priority = "high"
      class    = "blue"
    }
  }
}
```

### Add metrics configuration with S3 object filter for S3 Access Point

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_access_point" "example-access-point" {
  bucket = aws_s3_bucket.example.id
  name   = "example-access-point"
}

resource "aws_s3_bucket_metric" "example-filtered" {
  bucket = aws_s3_bucket.example.id
  name   = "ImportantBlueDocuments"

  filter {
    access_point = aws_s3_access_point.example-access-point.arn

    tags = {
      priority = "high"
      class    = "blue"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket` - (Required) Name of the bucket to put metric configuration.
* `name` - (Required) Unique identifier of the metrics configuration for the bucket. Must be less than or equal to 64 characters in length.
* `filter` - (Optional) [Object filtering](http://docs.aws.amazon.com/AmazonS3/latest/dev/metrics-configurations.html#metrics-configurations-filter) that accepts a prefix, tags, or a logical AND of prefix and tags (documented below).

The `filter` metric configuration supports the following:

~> **NOTE:** At least one of `access_point`, `prefix`, or `tags` is required when specifying a `filter`

* `access_point` - (Optional) S3 Access Point ARN for filtering (singular).
* `prefix` - (Optional) Object prefix for filtering (singular).
* `tags` - (Optional) Object tags for filtering (up to 10).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 bucket metric configurations using `bucket:metric`. For example:

```terraform
import {
  to = aws_s3_bucket_metric.my-bucket-entire-bucket
  id = "my-bucket:EntireBucket"
}
```

Using `terraform import`, import S3 bucket metric configurations using `bucket:metric`. For example:

```console
% terraform import aws_s3_bucket_metric.my-bucket-entire-bucket my-bucket:EntireBucket
```
