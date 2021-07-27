---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_intelligent_tiering_configuration"
description: |-
  Provides an S3 bucket intelligent tiering configuration resource.
---

# Resource: aws_s3_bucket_intelligent_tiering_configuration

Provides a S3 bucket [intelligent tiering configuration](https://docs.aws.amazon.com/AmazonS3/latest/userguide/storage-class-intro.html) resource.

## Example Usage

### Add intelligent tiering configuration for entire S3 bucket

```terraform
resource "aws_s3_bucket_intelligent_tiering_configuration" "example-entire-bucket" {
  bucket = aws_s3_bucket.example.bucket
  name   = "EntireBucket"

  tier {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 180
  }
  tier {
    access_tier = "ARCHIVE_ACCESS"
    days        = 125
  }

}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}

```

### Add intelligent tiering configuration with S3 bucket object filter

```terraform
resource "aws_s3_bucket_intelligent_tiering_configuration" "example-filtered" {
  bucket = aws_s3_bucket.example.bucket
  name   = "ImportantBlueDocuments"

  filter {
    prefix = "documents/"

    tags = {
      priority = "high"
      class    = "blue"
    }
  }

  tier {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 180
  }

  tier {
    access_tier = "ARCHIVE_ACCESS"
    days        = 125
  }

}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket this intelligent tiering configuration is associated with.
* `name` - (Required) Unique identifier of the intelligent tiering configuration for the bucket.
* `filter` - (Optional) Object filtering that accepts a prefix, tags, or a logical AND of prefix and tags (documented below).

The `filter` configuration supports the following:

* `prefix` - (Optional) Object prefix for filtering.
* `tags` - (Optional) Set of object tags for filtering.

## Attributes Reference

No additional attributes are exported.

## Import

S3 bucket analytics configurations can be imported using `???`, e.g.

```
$ terraform import aws_s3_bucket_intelligent_tiering_configuration.my-bucket-entire-bucket my-bucket:EntireBucket
```
