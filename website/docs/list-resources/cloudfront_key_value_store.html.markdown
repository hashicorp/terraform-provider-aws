---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_key_value_store"
description: |-
  Lists CloudFront Key Value Store resources.
---

# List Resource: aws_cloudfront_key_value_store

Lists CloudFront Key Value Store resources.

## Example Usage

```terraform
list "aws_cloudfront_key_value_store" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
