---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_analytics_configuration"
description: |-
  Provides a S3 bucket analytics configuration resource.
---

# Resource: aws_s3_bucket_analytics_configuration

Provides a S3 bucket [analytics configuration](https://docs.aws.amazon.com/AmazonS3/latest/dev/analytics-storage-class.html) resource.

-> This resource cannot be used with S3 directory buckets.

## Example Usage

### Add analytics configuration for entire S3 bucket and export results to a second S3 bucket

```terraform
resource "aws_s3_bucket_analytics_configuration" "example-entire-bucket" {
  bucket = aws_s3_bucket.example.id
  name   = "EntireBucket"

  storage_class_analysis {
    data_export {
      destination {
        s3_bucket_destination {
          bucket_arn = aws_s3_bucket.analytics.arn
        }
      }
    }
  }
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket" "analytics" {
  bucket = "analytics destination"
}
```

### Add analytics configuration with S3 object filter

```terraform
resource "aws_s3_bucket_analytics_configuration" "example-filtered" {
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

resource "aws_s3_bucket" "example" {
  bucket = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket` - (Required) Name of the bucket this analytics configuration is associated with.
* `name` - (Required) Unique identifier of the analytics configuration for the bucket.
* `filter` - (Optional) Object filtering that accepts a prefix, tags, or a logical AND of prefix and tags (documented below).
* `storage_class_analysis` - (Optional) Configuration for the analytics data export (documented below).

The `filter` configuration supports the following:

* `prefix` - (Optional) Object prefix for filtering.
* `tags` - (Optional) Set of object tags for filtering.

The `storage_class_analysis` configuration supports the following:

* `data_export` - (Required) Data export configuration (documented below).

The `data_export` configuration supports the following:

* `output_schema_version` - (Optional) Schema version of exported analytics data. Allowed values: `V_1`. Default value: `V_1`.
* `destination` - (Required) Specifies the destination for the exported analytics data (documented below).

The `destination` configuration supports the following:

* `s3_bucket_destination` - (Required) Analytics data export currently only supports an S3 bucket destination (documented below).

The `s3_bucket_destination` configuration supports the following:

* `bucket_arn` - (Required) ARN of the destination bucket.
* `bucket_account_id` - (Optional) Account ID that owns the destination bucket.
* `format` - (Optional) Output format of exported analytics data. Allowed values: `CSV`. Default value: `CSV`.
* `prefix` - (Optional) Prefix to append to exported analytics data.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 bucket analytics configurations using `bucket:analytics`. For example:

```terraform
import {
  to = aws_s3_bucket_analytics_configuration.my-bucket-entire-bucket
  id = "my-bucket:EntireBucket"
}
```

Using `terraform import`, import S3 bucket analytics configurations using `bucket:analytics`. For example:

```console
% terraform import aws_s3_bucket_analytics_configuration.my-bucket-entire-bucket my-bucket:EntireBucket
```
