---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_cors_configuration"
description: |-
  Lists S3 (Simple Storage) Bucket CORS Configuration resources.
---

# List Resource: aws_s3_bucket_cors_configuration

Lists S3 (Simple Storage) Bucket CORS Configuration resources for general purpose buckets with a CORS configuration.

## Example Usage

```terraform
list "aws_s3_bucket_cors_configuration" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
