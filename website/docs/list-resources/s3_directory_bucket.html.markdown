---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_directory_bucket"
description: |-
  Lists Amazon S3 Express Directory Bucket resources.
---

# List Resource: aws_s3_directory_bucket

Lists Amazon S3 Express Directory Bucket resources.

## Example Usage

```terraform
list "aws_s3_directory_bucket" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
