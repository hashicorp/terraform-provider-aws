---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_origin_request_policy"
description: |-
  Provides a CloudFront Origin Request Policy resource.
---

# Resource: aws_cloudfront_origin_request_policy

Provides a CloudFront Origin Request Policy resource, which determines the query string parameters, headers, and cookies that CloudFront forwards to an origin.

Read more about origin request policies in [Controlling origin requests](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/controlling-origin-requests.html).

## Example Usage

### Basic usage

```hcl
resource "aws_cloudfront_origin_request_policy" "example" {
  cookie_behavior       = "none"
  header_behavior       = "whitelist"
  header_names          = ["Authorization", "Host"]
  name                  = "Example-OriginRequestPolicy"
  query_string_behavior = "all"
}
```

## Argument Reference

The following arguments are supported:

* `comment` - (Optional) A comment to describe the origin request policy.

* `cookie_behavior` - (Required) Determines whether cookies in viewer requests are included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `all`. 

* `cookie_names` - (Optional) Specifies the cookies to be handled in accordance with the cookie behavior.

* `header_behavior` - (Required) Determines whether any HTTP headers are included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `allViewer`, `allViewerAndWhitelistCloudFront`.

* `header_names` - (Optional) Specifies the headers to be handled in accordance with the header behavior.

* `name` - (Required) A unique name to identify the cache policy.

* `query_string_behavior` - (Required) Determines whether any URL query strings in viewer requests are included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `all`.

* `query_string_names` - (Optional) Specifies the query string parameters to be handled in accordance with the query string behavior.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the origin request policy.
* `etag` - The ETag of the latest version of the origin request policy.

## Import

Cache Policies can be imported using the `id`, e.g.

```
$ terraform import aws_cloudfront_origin_request_policy.default 4ae686d6-72b3-4bb5-83f0-b89276902434
```
