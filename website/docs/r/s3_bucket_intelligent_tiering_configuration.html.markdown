---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_intelligent_tiering_configuration"
description: |-
  Provides an S3 Intelligent-Tiering configuration resource.
---

# Resource: aws_s3_bucket_intelligent_tiering_configuration

Provides an [S3 Intelligent-Tiering](https://docs.aws.amazon.com/AmazonS3/latest/userguide/intelligent-tiering.html) configuration resource.

## Example Usage

### Add intelligent tiering configuration for entire S3 bucket

```terraform
resource "aws_s3_bucket_intelligent_tiering_configuration" "example-entire-bucket" {
  bucket = aws_s3_bucket.example.id
  name   = "EntireBucket"

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 180
  }
  tiering {
    access_tier = "ARCHIVE_ACCESS"
    days        = 125
  }
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}
```

### Add intelligent tiering configuration with S3 object filter

```terraform
resource "aws_s3_bucket_intelligent_tiering_configuration" "example-filtered" {
  bucket = aws_s3_bucket.example.id
  name   = "ImportantBlueDocuments"

  status = "Disabled"

  filter {
    prefix = "documents/"

    tags = {
      priority = "high"
      class    = "blue"
    }
  }

  tiering {
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

* `bucket` - (Required) Name of the bucket this intelligent tiering configuration is associated with.
* `name` - (Required) Unique name used to identify the S3 Intelligent-Tiering configuration for the bucket.
* `status` - (Optional) Specifies the status of the configuration. Valid values: `Enabled`, `Disabled`.
* `filter` - (Optional) Bucket filter. The configuration only includes objects that meet the filter's criteria (documented below).
* `tiering` - (Required) S3 Intelligent-Tiering storage class tiers of the configuration (documented below).

The `filter` configuration supports the following:

* `prefix` - (Optional) Object key name prefix that identifies the subset of objects to which the configuration applies.
* `tags` - (Optional) All of these tags must exist in the object's tag set in order for the configuration to apply.

The `tiering` configuration supports the following:

* `access_tier` - (Required) S3 Intelligent-Tiering access tier. Valid values: `ARCHIVE_ACCESS`, `DEEP_ARCHIVE_ACCESS`.
* `days` - (Required) Number of consecutive days of no access after which an object will be eligible to be transitioned to the corresponding tier.

## Attributes Reference

No additional attributes are exported.

## Import

S3 bucket intelligent tiering configurations can be imported using `bucket:name`, e.g.

```
$ terraform import aws_s3_bucket_intelligent_tiering_configuration.my-bucket-entire-bucket my-bucket:EntireBucket
```
