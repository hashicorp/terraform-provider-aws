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

* `region` - (Optional) Region to query. Defaults to provider region.