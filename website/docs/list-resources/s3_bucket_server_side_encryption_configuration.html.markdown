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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
