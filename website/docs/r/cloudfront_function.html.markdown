---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_function"
description: |-
  Provides a CloudFront Function resource. With CloudFront Functions in Amazon CloudFront, you can write lightweight functions in JavaScript for high-scale, latency-sensitive CDN customizations.
---

# Resource: aws_cloudfront_function

Provides a CloudFront Function resource. With CloudFront Functions in Amazon CloudFront, you can write lightweight functions in JavaScript for high-scale, latency-sensitive CDN customizations.

See [CloudFront Functions](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/cloudfront-functions.html)

~> **NOTE:** You cannot delete a function if it’s associated with a cache behavior. First, update your distributions to remove the function association from all cache behaviors, then delete the function.

## Example Usage

### Basic Example

```terraform
resource "aws_cloudfront_function" "test" {
  name    = "test"
  runtime = "cloudfront-js-2.0"
  comment = "my function"
  publish = true
  code    = file("${path.module}/function.js")
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Unique name for your CloudFront Function.
* `code` - (Required) Source code of the function
* `runtime` - (Required) Identifier of the function's runtime. Valid values are `cloudfront-js-1.0` and `cloudfront-js-2.0`.

The following arguments are optional:

* `comment` - (Optional) Comment.
* `publish` - (Optional) Whether to publish creation/change as Live CloudFront Function Version. Defaults to `true`.
* `key_value_store_associations` - (Optional) List of `aws_cloudfront_key_value_store` ARNs to be associated to the function. AWS limits associations to on key value store per function.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifying your CloudFront Function.
* `etag` - ETag hash of the function. This is the value for the `DEVELOPMENT` stage of the function.
* `live_stage_etag` - ETag hash of any `LIVE` stage of the function.
* `status` - Status of the function. Can be `UNPUBLISHED`, `UNASSOCIATED` or `ASSOCIATED`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Functions using the `name`. For example:

```terraform
import {
  to = aws_cloudfront_function.test
  id = "my_test_function"
}
```

Using `terraform import`, import CloudFront Functions using the `name`. For example:

```console
% terraform import aws_cloudfront_function.test my_test_function
```
