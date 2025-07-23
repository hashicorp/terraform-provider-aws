---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_origin_request_policy"
description: |-
  Determines the values that CloudFront includes in requests that it sends to the origin.
---

# Resource: aws_cloudfront_origin_request_policy

## Example Usage

The following example below creates a CloudFront origin request policy.

```terraform
resource "aws_cloudfront_origin_request_policy" "example" {
  name    = "example-policy"
  comment = "example comment"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["example"]
    }
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["example"]
    }
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["example"]
    }
  }
}

```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Unique name to identify the origin request policy.
* `comment` - (Optional) Comment to describe the origin request policy.
* `cookies_config` - (Required) Object that determines whether any cookies in viewer requests (and if so, which cookies) are included in the origin request key and automatically included in requests that CloudFront sends to the origin. See [Cookies Config](#cookies-config) for more information.
* `headers_config` - (Required) Object that determines whether any HTTP headers (and if so, which headers) are included in the origin request key and automatically included in requests that CloudFront sends to the origin. See [Headers Config](#headers-config) for more information.
* `query_strings_config` - (Required) Object that determines whether any URL query strings in viewer requests (and if so, which query strings) are included in the origin request key and automatically included in requests that CloudFront sends to the origin. See [Query String Config](#query-string-config) for more information.

### Cookies Config

`cookie_behavior` - (Required) Determines whether any cookies in viewer requests are included in the origin request key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `all`, `allExcept`.
`cookies` - (Optional) Object that contains a list of cookie names. See [Items](#items) for more information.

### Headers Config

`header_behavior` - (Required) Determines whether any HTTP headers are included in the origin request key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `allViewer`, `allViewerAndWhitelistCloudFront`, `allExcept`.
`headers` - (Optional) Object that contains a list of header names. See [Items](#items) for more information.

### Query String Config

`query_string_behavior` - (Required) Determines whether any URL query strings in viewer requests are included in the origin request key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `all`, `allExcept`.
`query_strings` - (Optional) Object that contains a list of query string names. See [Items](#items) for more information.

### Items

`items` - (Required) List of item names (cookies, headers, or query strings).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The origin request policy ARN.
* `etag` - The current version of the origin request policy.
* `id` - The identifier for the origin request policy.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cloudfront Origin Request Policies using the `id`. For example:

```terraform
import {
  to = aws_cloudfront_origin_request_policy.policy
  id = "ccca32ef-dce3-4df3-80df-1bd3000bc4d3"
}
```

Using `terraform import`, import Cloudfront Origin Request Policies using the `id`. For example:

```console
% terraform import aws_cloudfront_origin_request_policy.policy ccca32ef-dce3-4df3-80df-1bd3000bc4d3
```
