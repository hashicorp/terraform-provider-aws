---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_origin_request_policy"
description: |-
  Determines the values that CloudFront includes in requests that it sends to the origin.
---

# Data Source: aws_cloudfront_origin_request_policy

## Example Usage

### Basic Usage

```terraform
data "aws_cloudfront_origin_request_policy" "example" {
  name = "example-policy"
}
```

### AWS-Managed Policies

AWS managed origin request policy names are prefixed with `Managed-`:

```terraform
data "aws_cloudfront_origin_request_policy" "ua_referer" {
  name = "Managed-UserAgentRefererHeaders"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - Unique name to identify the origin request policy.
* `id` - Identifier for the origin request policy.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `comment` - Comment to describe the origin request policy.
* `cookies_config` - Object that determines whether any cookies in viewer requests (and if so, which cookies) are included in the origin request key and automatically included in requests that CloudFront sends to the origin. See [Cookies Config](#cookies-config) for more information.
* `etag` - Current version of the origin request policy.
* `headers_config` - Object that determines whether any HTTP headers (and if so, which headers) are included in the origin request key and automatically included in requests that CloudFront sends to the origin. See [Headers Config](#headers-config) for more information.
* `query_strings_config` - Object that determines whether any URL query strings in viewer requests (and if so, which query strings) are included in the origin request key and automatically included in requests that CloudFront sends to the origin. See [Query String Config](#query-string-config) for more information.

### Cookies Config

`cookie_behavior` - Determines whether any cookies in viewer requests are included in the origin request key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist` `all`.
`cookies` - Object that contains a list of cookie names. See [Items](#items) for more information.

### Headers Config

`header_behavior` - Determines whether any HTTP headers are included in the origin request key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `allViewer`, `allViewerAndWhitelistCloudFront`.
`headers` - Object that contains a list of header names. See [Items](#items) for more information.

### Query String Config

`query_string_behavior` - Determines whether any URL query strings in viewer requests are included in the origin request key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `all`.
`query_strings` - Object that contains a list of query string names. See [Items](#items) for more information.

### Items

`items` - List of item names (cookies, headers, or query strings).
