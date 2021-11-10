---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_response_headers_policy"
description: |-
  Provides a CloudFront response headers policy resource.
---

# Resource: aws_cloudfront_response_headers_policy

Provides a CloudFront response headers policy resource.
A response headers policy contains information about a set of HTTP response headers and their values.
After you create a response headers policy, you can use its ID to attach it to one or more cache behaviors in a CloudFront distribution.
When it’s attached to a cache behavior, CloudFront adds the headers in the policy to every response that it sends for requests that match the cache behavior.

## Example Usage

The following example below creates a CloudFront response headers policy.

```terraform
resource "aws_cloudfront_response_headers_policy" "example" {
  name    = "example-policy"
  comment = "test comment"

  cors_config {
    access_control_allow_credentials = true

    access_control_allow_headers {
      items = ["test"]
    }

    access_control_allow_methods {
      items = ["GET"]
    }

    access_control_allow_origins {
      items = ["test.example.comtest"]
    }

    origin_override = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name to identify the response headers policy.
* `comment` - (Optional) A comment to describe the response headers policy. The comment cannot be longer than 128 characters.
* `cors_config` - (Optional) A configuration for a set of HTTP response headers that are used for Cross-Origin Resource Sharing (CORS). See [Cors Config](#cors_config) for more information.
* `custom_headers_config` - (Optional) Object that contains an attribute `items` that contains a list of Custom Headers See [Custom Header](#custom_header) for more information.
* `security_headers_config` - (Optional) A configuration for a set of security-related HTTP response headers. See [Security Headers Config](#security_headers_config) for more information.

### Cors Config

* `access_control_allow_credentials` - (Required) A Boolean value that CloudFront uses as the value for the Access-Control-Allow-Credentials HTTP response header.
* `access_control_allow_headers` - (Required) Object that contains an attribute `items` that contains a list of HTTP header names that CloudFront includes as values for the Access-Control-Allow-Headers HTTP response header.
* `access_control_allow_methods` - (Required) Object that contains an attribute `items` that contains a list of HTTP methods that CloudFront includes as values for the Access-Control-Allow-Methods HTTP response header. Valid values: `GET` | `POST` | `OPTIONS` | `PUT` | `DELETE` | `HEAD` | `ALL`
* `access_control_allow_origins` - (Optional) Object that contains an attribute `items` that contains a list of origins that CloudFront can use as the value for the Access-Control-Allow-Origin HTTP response header.
* `access_control_expose_headers` - (Optional) Object that contains an attribute `items` that contains a list of HTTP headers that CloudFront includes as values for the Access-Control-Expose-Headers HTTP response header.
* `access_control_max_age_sec` - (Required) A number that CloudFront uses as the value for the Access-Control-Max-Age HTTP response header.

### Custom Header

* `header` - (Required) The HTTP response header name.
* `override` - (Required) A Boolean value that determines whether CloudFront overrides a response header with the same name received from the origin with the header specifies here.
* `value` - (Required) The value for the HTTP response header.

### Security Headers Config

* `content_security_policy` - (Optional) The policy directives and their values that CloudFront includes as values for the Content-Security-Policy HTTP response header. See [Content Security Policy](#content_security_policy) for more information.
* `content_type_options` - (Optional) TA setting that determines whether CloudFront includes the X-Content-Type-Options HTTP response header with its value set to nosniff. See [Content Type Options](#content_type_options) for more information.
* `frame_options` - (Optional) TA setting that determines whether CloudFront includes the X-Frame-Options HTTP response header and the header’s value. See [Frame Options](#frame_options) for more information.
* `referrer_policy` - (Optional) TA setting that determines whether CloudFront includes the Referrer-Policy HTTP response header and the header’s value. See [Referrer Policy](#referrer_policy) for more information.
* `xss_protection` - (Optional) TSettings that determine whether CloudFront includes the X-XSS-Protection HTTP response header and the header’s value. See [XSS Protection](#xss_protection) for more information.

### Content Security Policy

* `content_security_policy` - (Required) TThe policy directives and their values that CloudFront includes as values for the Content-Security-Policy HTTP response header.
* `override` - (Required) A Boolean value that determines whether CloudFront overrides the Content-Security-Policy HTTP response header received from the origin with the one specified in this response headers policy.

### Content Type Options

* `override` - (Required) A Boolean value that determines whether CloudFront overrides the X-Content-Type-Options HTTP response header received from the origin with the one specified in this response headers policy.

### Frame Options

* `frame_option` - (Required) The value of the X-Frame-Options HTTP response header. Valid values: `DENY` | `SAMEORIGIN`
* `override` - (Required) A Boolean value that determines whether CloudFront overrides the X-Frame-Options HTTP response header received from the origin with the one specified in this response headers policy.

### Referrer Policy

* `referrer_policy` - (Required) The value of the Referrer-Policy HTTP response header. Valid Values: `no-referrer` | `no-referrer-when-downgrade` | `origin` | `origin-when-cross-origin` | `same-origin` | `strict-origin` | `strict-origin-when-cross-origin` | `unsafe-url`
* `override` - (Required) A Boolean value that determines whether CloudFront overrides the Referrer-Policy HTTP response header received from the origin with the one specified in this response headers policy.

### Strict Transport Security

* `access_control_max_age_sec` - (Required) A number that CloudFront uses as the value for the max-age directive in the Strict-Transport-Security HTTP response header.
* `include_subdomains` - (Optional) A Boolean value that determines whether CloudFront includes the includeSubDomains directive in the Strict-Transport-Security HTTP response header.
* `override` - (Required) A Boolean value that determines whether CloudFront overrides the Strict-Transport-Security HTTP response header received from the origin with the one specified in this response headers policy.
* `preload` - (Optional) A Boolean value that determines whether CloudFront includes the preload directive in the Strict-Transport-Security HTTP response header.

### XSS Protection

* `mode_block` - (Required) A Boolean value that determines whether CloudFront includes the mode=block directive in the X-XSS-Protection header.
* `override` - (Required) A Boolean value that determines whether CloudFront overrides the X-XSS-Protection HTTP response header received from the origin with the one specified in this response headers policy.
* `protection` - (Required) A Boolean value that determines the value of the X-XSS-Protection HTTP response header. When this setting is true, the value of the X-XSS-Protection header is 1. When this setting is false, the value of the X-XSS-Protection header is 0.
* `report_uri` - (Optional) A Boolean value that determines whether CloudFront sets a reporting URI in the X-XSS-Protection header.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `etag` - The current version of the response headers policy.
* `id` - The identifier for the response headers policy.

## Import

Cloudfront Response Headers Policies can be imported using the `id`, e.g.

```
$ terraform import aws_cloudfront_response_headers_policy.policy 658327ea-f89d-4fab-a63d-7e88639e58f9
```