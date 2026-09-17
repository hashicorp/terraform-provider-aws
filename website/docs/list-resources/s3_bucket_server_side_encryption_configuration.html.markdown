---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_server_side_encryption_configuration"
description: |-
  Lists S3 (Simple Storage) Bucket Server Side Encryption Configuration resources.
---

# List Resource: aws_s3_bucket_server_side_encryption_configuration

Lists S3 (Simple Storage) Bucket Server Side Encryption Configuration resources.

## Example Usage

```terraform
list "aws_s3_bucket_server_side_encryption_configuration" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
