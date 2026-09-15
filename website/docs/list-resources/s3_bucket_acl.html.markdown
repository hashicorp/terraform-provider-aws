---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_acl"
description: |-
  Lists S3 (Simple Storage) Bucket ACL resources.
---

# List Resource: aws_s3_bucket_acl

Lists S3 (Simple Storage) Bucket ACL resources.

## Example Usage

```terraform
list "aws_s3_bucket_acl" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
