---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_cache_policy"
description: |-
  Use the `aws_cloudfront_cache_policy` resource to manage cache policies for CloudFront distributions. When attached to a cache behavior, the cache policy determines the values included in the cache key, such as HTTP headers, cookies, and URL query strings. CloudFront uses the cache key to locate cached objects to return to viewers. The cache policy also determines the default, minimum, and maximum time to live (TTL) values for objects in the CloudFront cache.
---

# Resource: aws_cloudfront_cache_policy

## Example Usage

Use the `aws_cloudfront_cache_policy` resource to create a CloudFront cache policy.

```terraform
resource "aws_cloudfront_cache_policy" "example" {
  name        = "example-policy"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
  parameters_in_cache_key_and_forwarded_to_origin {
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
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Unique name for identifying the cache policy.
* `min_ttl` - (Required) Minimum amount of time, in seconds, that you want objects to stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated.
* `max_ttl` - (Optional) Maximum amount of time, in seconds, that objects stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated.
* `default_ttl` - (Optional) Default time, in seconds, that you want objects to stay in the CloudFront cache before CloudFront sends another request to the origin to check if the object has been updated.
* `comment` - (Optional) Comment that describes the cache policy.
* `parameters_in_cache_key_and_forwarded_to_origin` - (Required) Configuration for including HTTP headers, cookies, and URL query strings in the cache key. For more information, refer to the [Parameters In Cache Key And Forwarded To Origin](#parameters-in-cache-key-and-forwarded-to-origin) section.

### Parameters In Cache Key And Forwarded To Origin

* `cookies_config` - (Required) Object that determines whether any cookies in viewer requests (and if so, which cookies) are included in the cache key and automatically included in requests that CloudFront sends to the origin. See [Cookies Config](#cookies-config) for more information.
* `headers_config` - (Required) Object that determines whether any HTTP headers, and if so, which headers, are included in the cache key and automatically included in requests that CloudFront sends to the origin. See [Headers Config](#headers-config) for more information.
* `query_strings_config` - (Required) Object that determines whether any URL query strings in viewer requests, and if so, which query strings, are included in the cache key. It also automatically includes these query strings in requests that CloudFront sends to the origin. Please refer to the [Query String Config](#query-string-config) for more information.
* `enable_accept_encoding_brotli` - (Optional) Flag that can affect whether the Accept-Encoding HTTP header is included in the cache key and included in requests that CloudFront sends to the origin.
* `enable_accept_encoding_gzip` - (Optional) Flag that affects whether the Accept-Encoding HTTP header is included in the cache key and included in requests that CloudFront sends to the origin.

### Cookies Config

* `cookie_behavior` - (Required) Whether any cookies in viewer requests are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values for `cookie_behavior` are `none`, `whitelist`, `allExcept`, and `all`.
* `cookies` - (Optional) Object that contains a list of cookie names. See [Items](#items) for more information.

### Headers Config

* `header_behavior` - (Required) Whether any HTTP headers are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values for `header_behavior` are `none` and `whitelist`.
* `headers` - (Optional) Object contains a list of header names. See [Items](#items) for more information.

### Query String Config

* `query_string_behavior` - (Required) Whether URL query strings in viewer requests are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values for `query_string_behavior` are `none`, `whitelist`, `allExcept`, and `all`.
* `query_strings` - (Optional) Object that contains a list of query string names. See [Items](#items) for more information.

### Items

* `items` - (Required) List of item names, such as cookies, headers, or query strings.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `etag` - Current version of the cache policy.
* `id` - Identifier for the cache policy.

## Import

CloudFront cache policies can be imported using their `id`. For example,

```
$ terraform import aws_cloudfront_cache_policy.policy 658327ea-f89d-4fab-a63d-7e88639e58f6
```
