---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_cache_policy"
description: |-
  Provides a CloudFront Cache Policy resource.
---

# Resource: aws_cloudfront_cache_policy

Provides a CloudFront Cache Policy resource, which allows for fine-grained control of the query string parameters, headers, and cookies that are included in the cache key by CloudFront.

Read more about cache policies in [Controlling the cache key](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/controlling-the-cache-key.html).

## Example Usage

### Basic usage

```hcl
resource "aws_cloudfront_cache_policy" "example" {
  default_ttl     = 3600
  header_behavior = "whitelist"
  header_names    = ["Authorization", "Host"]
  max_ttl         = 86400
  min_ttl         = 0
  name            = "Example-CachePolicy"
}
```

## Argument Reference

The following arguments are supported:

* `comment` - (Optional) A comment to describe the cache policy.

* `cookie_behavior` - (Optional) Determines whether any cookies in viewer requests are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values are `none` (default), `whitelist`, `allExcept`, `all`. 

* `cookie_names` - (Optional) Specifies the cookies to be handled in accordance with the cookie behavior.
 
* `default_ttl` - (Optional) The default amount of time, in seconds, that you want objects to stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated. CloudFront uses this value as the object’s time to live (TTL) only when the origin does not send `Cache-Control` or `Expires` headers with the object. [See here for more information.](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/Expiration.html) The default value for this field is 86400 seconds (1 day). If the value of `min_ttl` is more than 86400 seconds, then the default value for this field is the same as the value of `min_ttl`. 

* `enable_accept_encoding_brotli` - (Optional) A flag that determines whether the `Accept-Encoding` HTTP header is included in the cache key and included in requests that CloudFront sends to the origin. [See here for more information](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/controlling-the-cache-key.html#cache-policy-compressed-objects). The default value of this field is `false`.

* `enable_accept_encoding_gzip` - (Optional) A flag that determines whether the `Accept-Encoding` HTTP header is included in the cache key and included in requests that CloudFront sends to the origin. [See here for more information](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/controlling-the-cache-key.html#cache-policy-compressed-objects). The default value of this field is `false`.

* `header_behavior` - (Optional) Determines whether any HTTP headers are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values are `none` (default) and `whitelist`.

* `header_names` - (Optional) Specifies the headers to be handled in accordance with the header behavior.

* `max_ttl` - (Optional) The maximum amount of time, in seconds, that objects stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated. CloudFront uses this value only when the origin sends `Cache-Control` or `Expires` headers with the object. [See here for more information.](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/Expiration.html) The default value for this field is 31536000 seconds (1 year). If the value of `min_ttl` or `default_ttl` is more than 31536000 seconds, then the default value for this field is the same as the value of `default_ttl`.

* `min_ttl` - (Required) The minimum amount of time, in seconds, that you want objects to stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated. [See here for more information.](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/Expiration.html)

* `name` - (Required) A unique name to identify the cache policy.

* `query_string_behavior` - (Optional) Determines whether any URL query strings in viewer requests are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values are `none` (default), `whitelist`, `allExcept`, `all`.

* `query_string_names` - (Optional) Specifies the query string parameters to be handled in accordance with the query string behavior.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the cache policy.
* `etag` - The ETag of the latest version of the cache policy.

## Import

Cache Policies can be imported using the `id`, e.g.

```
$ terraform import aws_cloudfront_cache_policy.default 486631b3-a60c-413e-82c4-f445871fb970
```
