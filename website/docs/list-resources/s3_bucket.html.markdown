---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket"
description: |-
  Lists S3 (Simple Storage) Bucket resources.
---

# List Resource: aws_s3_bucket

Lists S3 (Simple Storage) Bucket resources.

## Example Usage

```terraform
list "aws_s3_bucket" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
