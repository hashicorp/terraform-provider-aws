---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_lifecycle_configuration"
description: |-
  Lists S3 (Simple Storage) Bucket Lifecycle Configuration resources.
---

# List Resource: aws_s3_bucket_lifecycle_configuration

Lists S3 (Simple Storage) Bucket Lifecycle Configuration resources.

## Example Usage

```terraform
list "aws_s3_bucket_lifecycle_configuration" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
