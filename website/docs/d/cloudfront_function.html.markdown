---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_function"
description: |-
  Provides a CloudFront Function data source.
---

# aws_cloudfront_function

Provides information about a CloudFront Function.

## Example Usage

```terraform
variable "function_name" {
  type = string
}

data "aws_cloudfront_function" "existing" {
  name = var.function_name
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the CloudFront function.
* `stage` - (Required) Function’s stage, either `DEVELOPMENT` or `LIVE`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN identifying your CloudFront Function.
* `code` - Source code of the function
* `comment` - Comment.
* `etag` - ETag hash of the function
* `key_value_store_associations` - List of `aws_cloudfront_key_value_store` ARNs associated to the function.
* `last_modified_time` - When this resource was last modified.
* `runtime` - Identifier of the function's runtime.
* `status` - Status of the function. Can be `UNPUBLISHED`, `UNASSOCIATED` or `ASSOCIATED`.
