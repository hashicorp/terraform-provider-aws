---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_response_headers_policy"
description: |-
  Use this data source to retrieve information about a CloudFront response headers policy.
---

# Data source: aws_cloudfront_response_headers_policy

Use this data source to retrieve information about a CloudFront cache policy.

## Example Usage

```terraform
data "aws_cloudfront_response_headers_policy" "example" {
  name = "example-policy"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) A unique name to identify the response headers policy.
* `id` - (Optional) The identifier for the response headers policy.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `comment` - A comment to describe the response headers policy. The comment cannot be longer than 128 characters.
* `etag` - The current version of the response headers policy.
* `cors_config` - A configuration for a set of HTTP response headers that are used for Cross-Origin Resource Sharing (CORS). See [Cors Config](#cors-config) for more information.
* `custom_headers_config` - Object that contains an attribute `items` that contains a list of Custom Headers See [Custom Header](#custom-header) for more information.
* `security_headers_config` - A configuration for a set of security-related HTTP response headers. See [Security Headers Config](#security-headers-config) for more information.
* `server_timing_headers_config` - (Optional) A configuration for enabling the Server-Timing header in HTTP responses sent from CloudFront. See [Server Timing Headers Config](#server-timing-headers-config) for more information.

### Cors Config

* `access_control_allow_credentials` - A Boolean value that CloudFront uses as the value for the Access-Control-Allow-Credentials HTTP response header.
* `access_control_allow_headers` - Object that contains an attribute `items` that contains a list of HTTP header names that CloudFront includes as values for the Access-Control-Allow-Headers HTTP response header.
* `access_control_allow_methods` - Object that contains an attribute `items` that contains a list of HTTP methods that CloudFront includes as values for the Access-Control-Allow-Methods HTTP response header. Valid values: `GET` | `POST` | `OPTIONS` | `PUT` | `DELETE` | `HEAD` | `ALL`
* `access_control_allow_origins` - Object that contains an attribute `items` that contains a list of origins that CloudFront can use as the value for the Access-Control-Allow-Origin HTTP response header.
* `access_control_expose_headers` - Object that contains an attribute `items` that contains a list of HTTP headers that CloudFront includes as values for the Access-Control-Expose-Headers HTTP response header.
* `access_control_max_age_sec` - A number that CloudFront uses as the value for the Access-Control-Max-Age HTTP response header.

### Custom Header

* `header` - The HTTP response header name.
* `override` - A Boolean value that determines whether CloudFront overrides a response header with the same name received from the origin with the header specifies here.
* `value` - The value for the HTTP response header.

### Security Headers Config

* `content_security_policy` - The policy directives and their values that CloudFront includes as values for the Content-Security-Policy HTTP response header. See [Content Security Policy](#content-security-policy) for more information.
* `content_type_options` - A setting that determines whether CloudFront includes the X-Content-Type-Options HTTP response header with its value set to nosniff. See [Content Type Options](#content-type-options) for more information.
* `frame_options` - A setting that determines whether CloudFront includes the X-Frame-Options HTTP response header and the header’s value. See [Frame Options](#frame-options) for more information.
* `referrer_policy` - A setting that determines whether CloudFront includes the Referrer-Policy HTTP response header and the header’s value. See [Referrer Policy](#referrer-policy) for more information.
* `strict_transport_security` - Settings that determine whether CloudFront includes the Strict-Transport-Security HTTP response header and the header’s value. See [Strict Transport Security](#strict-transport-security) for more information.
* `xss_protection` - Settings that determine whether CloudFront includes the X-XSS-Protection HTTP response header and the header’s value. See [XSS Protection](#xss-protection) for more information.

### Content Security Policy

* `content_security_policy` - The policy directives and their values that CloudFront includes as values for the Content-Security-Policy HTTP response header.
* `override` - A Boolean value that determines whether CloudFront overrides the Content-Security-Policy HTTP response header received from the origin with the one specified in this response headers policy.

### Content Type Options

* `override` - A Boolean value that determines whether CloudFront overrides the X-Content-Type-Options HTTP response header received from the origin with the one specified in this response headers policy.

### Frame Options

* `frame_option` - The value of the X-Frame-Options HTTP response header. Valid values: `DENY` | `SAMEORIGIN`
* `override` - A Boolean value that determines whether CloudFront overrides the X-Frame-Options HTTP response header received from the origin with the one specified in this response headers policy.

### Referrer Policy

* `referrer_policy` - The value of the Referrer-Policy HTTP response header. Valid Values: `no-referrer` | `no-referrer-when-downgrade` | `origin` | `origin-when-cross-origin` | `same-origin` | `strict-origin` | `strict-origin-when-cross-origin` | `unsafe-url`
* `override` - A Boolean value that determines whether CloudFront overrides the Referrer-Policy HTTP response header received from the origin with the one specified in this response headers policy.

### Strict Transport Security

* `access_control_max_age_sec` - A number that CloudFront uses as the value for the max-age directive in the Strict-Transport-Security HTTP response header.
* `include_subdomains` - A Boolean value that determines whether CloudFront includes the includeSubDomains directive in the Strict-Transport-Security HTTP response header.
* `override` - A Boolean value that determines whether CloudFront overrides the Strict-Transport-Security HTTP response header received from the origin with the one specified in this response headers policy.
* `preload` - A Boolean value that determines whether CloudFront includes the preload directive in the Strict-Transport-Security HTTP response header.

### XSS Protection

* `mode_block` - A Boolean value that determines whether CloudFront includes the mode=block directive in the X-XSS-Protection header.
* `override` - A Boolean value that determines whether CloudFront overrides the X-XSS-Protection HTTP response header received from the origin with the one specified in this response headers policy.
* `protection` - A Boolean value that determines the value of the X-XSS-Protection HTTP response header. When this setting is true, the value of the X-XSS-Protection header is 1. When this setting is false, the value of the X-XSS-Protection header is 0.
* `report_uri` - A Boolean value that determines whether CloudFront sets a reporting URI in the X-XSS-Protection header.

### Server Timing Headers Config

* `enabled` - A Boolean that determines whether CloudFront adds the `Server-Timing` header to HTTP responses that it sends in response to requests that match a cache behavior that's associated with this response headers policy.
* `sampling_rate` - A number 0–100 (inclusive) that specifies the percentage of responses that you want CloudFront to add the Server-Timing header to.
