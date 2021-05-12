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

The following arguments are supported:

* `name` - (Required) Name of the CloudFront function.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `code` - Source code of the function
* `runtime` - Identifier of the function's runtime.
* `comment` - Comment.
* `publish` - Whether to publish creation/change as Live CloudFront Function Version.
* `arn` - Amazon Resource Name (ARN) identifying your CloudFront Function.
* `version` - ETag hash of the function
* `last_modified` - Date this resource was last modified.
* `status` - Status of the function. Can be `UNPUBLISHED`, `UNASSOCIATED` or `ASSOCIATED`.
* `stage` - Stage of the code. Can be `DEVELOPMENT` or `LIVE`.
