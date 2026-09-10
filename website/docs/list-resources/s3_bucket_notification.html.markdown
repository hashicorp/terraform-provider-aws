---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_notification"
description: |-
  Lists S3 (Simple Storage) Bucket Notification resources.
---

# List Resource: aws_s3_bucket_notification

Lists S3 (Simple Storage) Bucket Notification resources.

## Example Usage

### Basic Usage

```terraform
list "aws_s3_bucket_notification" "example" {
  provider = aws
}
```

### Include Resource Data

```terraform
list "aws_s3_bucket_notification" "example" {
  provider = aws

  include_resource = true
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
