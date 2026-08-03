---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_public_access_block"
description: |-
  Lists S3 (Simple Storage) Bucket Public Access Block resources.
---

# List Resource: aws_s3_bucket_public_access_block

Lists S3 (Simple Storage) Bucket Public Access Block resources.

## Example Usage

```terraform
list "aws_s3_bucket_public_access_block" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
