---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_ownership_controls"
description: |-
  Lists S3 (Simple Storage) Bucket Ownership Controls resources.
---

# List Resource: aws_s3_bucket_ownership_controls

Lists S3 (Simple Storage) Ownership Controls resources.

## Example Usage

```terraform
list "aws_s3_bucket_ownership_controls" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
