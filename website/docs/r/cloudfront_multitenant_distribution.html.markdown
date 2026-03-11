---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_multitenant_distribution"
description: |-
  Provides a CloudFront multi-tenant distribution resource.
---

# Resource: aws_cloudfront_multitenant_distribution

Creates an Amazon CloudFront multi-tenant distribution.

Multi-tenant distributions are a specialized type of CloudFront distribution designed for multi-tenant applications. They have specific limitations and requirements compared to standard CloudFront distributions.

For information about CloudFront multi-tenant distributions, see the [Amazon CloudFront Developer Guide](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/).

~> **NOTE:** CloudFront distributions take about 15 minutes to reach a deployed state after creation or modification. During this time, deletes to resources will be blocked. If you need to delete a distribution that is enabled and you do not want to wait, you need to use the `retain_on_delete` flag.

## Multi-tenant Distribution Limitations

Multi-tenant distributions have the following limitations compared to standard CloudFront distributions:

- **Connection Mode**: Automatically set to `tenant-only` and cannot be modified
- **Cache Policies**: Must use cache policies instead of legacy TTL settings
- **Trusted Key Groups**: Must use trusted key groups instead of trusted signers
- **WAF Integration**: Only supports WAF v2 web ACLs
- **Certificate Management**: Must use ACM certificates (IAM certificates not supported)

### Unsupported Attributes

The following attributes that are available in standard CloudFront distributions are **not supported** for multi-tenant distributions:

- `active_trusted_signers` - Use `active_trusted_key_groups` instead
- `alias_icp_recordals` - Managed by connection groups
- `aliases` - Managed by connection groups
- `anycast_ip_list_id` - Use connection groups instead
- `continuous_deployment_policy_id`
- `forwarded_values` in cache behaviors - Deprecated, use cache policies instead
- `is_ipv6_enabled` - Managed by connection groups
- `price_class` - Managed by connection groups
- `smooth_streaming` in cache behaviors
- `staging` mode
- `trusted_signers` in cache behaviors - Use `trusted_key_groups` instead
- Cache behavior TTL settings (`default_ttl`, `max_ttl`, `min_ttl`) - Use cache policies instead

## Example Usage

```terraform
resource "aws_cloudfront_multitenant_distribution" "example" {
  comment = "Multi-tenant distribution for my application"
  enabled = true

  origin {
    domain_name = "example.com"
    id          = "example-origin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    target_origin_id       = "example-origin"
    viewer_protocol_policy = "redirect-to-https"
    cache_policy_id        = aws_cloudfront_cache_policy.example.id

    allowed_methods {
      items          = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
      cached_methods = ["GET", "HEAD"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = aws_acm_certificate.example.arn
    ssl_support_method  = "sni-only"
  }

  tenant_config {
    parameter_definition {
      name = "origin_domain"
      definition {
        string_schema {
          required = true
          comment  = "Origin domain parameter for tenants"
        }
      }
    }
  }

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `cache_behavior` - (Optional) Ordered list of cache behaviors resource for this distribution. See [Cache Behavior](#cache-behavior) below.
* `comment` - (Required) Any comments you want to include about the distribution.
* `custom_error_response` - (Optional) One or more custom error response elements. See [Custom Error Response](#custom-error-response) below.
* `default_cache_behavior` - (Required) Default cache behavior for this distribution. See [Default Cache Behavior](#default-cache-behavior) below.
* `default_root_object` - (Optional) Object that you want CloudFront to return when an end user requests the root URL.
* `enabled` - (Required) Whether the distribution is enabled to accept end user requests for content.
* `http_version` - (Optional) Maximum HTTP version to support on the distribution. Allowed values are `http1.1`, `http2`, `http2and3`, and `http3`. Default: `http2`.
* `origin_group` - (Optional) One or more origin_group for this distribution (multiples allowed). See [Origin Group](#origin-group) below.
* `origin` - (Required) One or more origins for this distribution (multiples allowed). See [Origin](#origin) below.
* `restrictions` - (Required) Restriction configuration for this distribution. See [Restrictions](#restrictions) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tenant_config` - (Required) Tenant configuration that contains parameter definitions for multi-tenant distributions. See [Tenant Config](#tenant-config) below.
* `viewer_certificate` - (Required) SSL configuration for this distribution. See [Viewer Certificate](#viewer-certificate) below.
* `web_acl_id` - (Optional) Unique identifier that specifies the AWS WAF v2 web ACL to associate with this distribution.

### Cache Behavior

Cache behavior supports all the same arguments as [Default Cache Behavior](#default-cache-behavior) with the addition of:

* `path_pattern` - (Required) Pattern that specifies which requests you want this cache behavior to apply to.

### Default Cache Behavior

* `target_origin_id` - (Required) Value of ID for the origin that you want CloudFront to route requests to when a request matches the path pattern either for a cache behavior or for the default cache behavior.
* `viewer_protocol_policy` - (Required) Use this element to specify the protocol that users can use to access the files in the origin specified by TargetOriginId when a request matches the path pattern in PathPattern. One of `allow-all`, `https-only`, or `redirect-to-https`.
* `allowed_methods` - (Required) Controls which HTTP methods CloudFront processes and forwards to your Amazon S3 bucket or your custom origin.
* `cached_methods` - (Required) Controls whether CloudFront caches the response to requests using the specified HTTP methods.
* `cache_policy_id` - (Optional) Unique identifier of the cache policy that is attached to the cache behavior.
* `compress` - (Optional) Whether you want CloudFront to automatically compress content for web requests that include `Accept-Encoding: gzip` in the request header. Default: `false`.
* `field_level_encryption_id` - (Optional) Field level encryption configuration ID.
* `function_association` - (Optional) Configuration block for CloudFront Functions associations. See [Function Association](#function-association) below.
* `lambda_function_association` - (Optional) Configuration block for Lambda@Edge associations. See [Lambda Function Association](#lambda-function-association) below.
* `origin_request_policy_id` - (Optional) Unique identifier of the origin request policy that is attached to the behavior.
* `realtime_log_config_arn` - (Optional) ARN of the real-time log configuration that is attached to this cache behavior.
* `response_headers_policy_id` - (Optional) Identifier for a response headers policy.
* `trusted_key_groups` - (Optional) List of key group IDs that CloudFront can use to validate signed URLs or signed cookies.

### Function Association

* `event_type` - (Required) Specific event to trigger this function. Valid values: `viewer-request`, `origin-request`, `viewer-response`, `origin-response`.
* `function_arn` - (Required) ARN of the CloudFront function.

### Lambda Function Association

* `event_type` - (Required) Specific event to trigger this function. Valid values: `viewer-request`, `origin-request`, `viewer-response`, `origin-response`.
* `include_body` - (Optional) When set to true, the request body is exposed to the Lambda function. Default: `false`.
* `lambda_function_arn` - (Required) ARN of the Lambda function.

### Origin

* `domain_name` - (Required) DNS domain name of either the S3 bucket, or web site of your custom origin.
* `origin_id` - (Required) Unique identifier for the origin.
* `connection_attempts` - (Optional) Number of times that CloudFront attempts to connect to the origin. Must be between 1-3. Default: 3.
* `connection_timeout` - (Optional) Number of seconds that CloudFront waits when trying to establish a connection to the origin. Must be between 1-10. Default: 10.
* `custom_header` - (Optional) One or more sub-resources with `name` and `value` parameters that specify header data that will be sent to the origin. See [Custom Header](#custom-header) below.
* `custom_origin_config` - (Optional) CloudFront origin access identity to associate with the origin. See [Custom Origin Config](#custom-origin-config) below.
* `origin_access_control_id` - (Optional) CloudFront origin access control identifier to associate with the origin.
* `origin_path` - (Optional) Optional element that causes CloudFront to request your content from a directory in your Amazon S3 bucket or your custom origin.
* `origin_shield` - (Optional) CloudFront Origin Shield configuration information. See [Origin Shield](#origin-shield) below.
* `response_completion_timeout` - (Optional) Number of seconds that CloudFront waits for a response after forwarding a request to the origin. Default: 30.
* `vpc_origin_config` - (Optional) CloudFront VPC origin configuration. See [VPC Origin Config](#vpc-origin-config) below.

### Custom Header

* `header_name` - (Required) Name of the header.
* `header_value` - (Required) Value for the header.

### Custom Origin Config

* `http_port` - (Required) HTTP port the custom origin listens on.
* `https_port` - (Required) HTTPS port the custom origin listens on.
* `ip_address_type` - (Optional) Type of IP addresses used by your origins. Valid values are `ipv4` and `dualstack`.
* `origin_keepalive_timeout` - (Optional) Custom keep-alive timeout, in seconds. Default: 5.
* `origin_read_timeout` - (Optional) Custom read timeout, in seconds. Default: 30.
* `origin_protocol_policy` - (Required) Origin protocol policy to apply to your origin. Valid values are `http-only`, `https-only`, and `match-viewer`.
* `origin_ssl_protocols` - (Required) List of SSL/TLS protocols that you want CloudFront to use when communicating with your origin over HTTPS.

### Origin Shield

* `enabled` - (Required) Whether Origin Shield is enabled.
* `origin_shield_region` - (Optional) AWS Region for Origin Shield. Required when `enabled` is `true`.

### Origin Group

* `origin_id` - (Required) Unique identifier for the origin group.
* `failover_criteria` - (Required) Failover criteria for when to failover to the secondary origin. See [Failover Criteria](#failover-criteria) below.
* `member` - (Required) List of origins in this origin group. Must contain exactly 2 members. See [Origin Group Member](#origin-group-member) below.

### Failover Criteria

* `status_codes` - (Required) List of HTTP status codes that trigger a failover to the secondary origin.

### Origin Group Member

* `origin_id` - (Required) Unique identifier of an origin in the origin group.

### Restrictions

* `geo_restriction` - (Required) Geographic restriction configuration. See [Geo Restriction](#geo-restriction) below.

### Geo Restriction

* `restriction_type` - (Required) Method to restrict distribution of your content by country. Valid values are `none`, `whitelist`, and `blacklist`.
* `items` - (Optional) List of ISO 3166-1-alpha-2 country codes for which you want CloudFront either to distribute your content (`whitelist`) or not distribute your content (`blacklist`). Required when `restriction_type` is `whitelist` or `blacklist`.

### Active Trusted Key Groups

* `enabled` - Whether any of the key groups have public keys that CloudFront can use to verify the signatures of signed URLs and signed cookies.
* `items` - List of key groups. See [Key Group Items](#key-group-items) below.

### Key Group Items

* `key_group_id` - ID of the key group that contains the public keys.
* `key_pair_ids` - Set of active CloudFront key pairs associated with the signer that can be used to verify the signatures of signed URLs and signed cookies.

### VPC Origin Config

* `origin_keepalive_timeout` - (Optional) Custom keep-alive timeout, in seconds. By default, CloudFront uses a default timeout. Default: 5.
* `origin_read_timeout` - (Optional) Custom read timeout, in seconds. By default, CloudFront uses a default timeout. Default: 30.
* `vpc_origin_id` - (Required) ID of the VPC origin that you want CloudFront to route requests to.

### Custom Error Response

* `error_caching_min_ttl` - (Optional) Minimum amount of time that you want CloudFront to cache the HTTP status code specified in ErrorCode.
* `error_code` - (Required) HTTP status code for which you want to specify a custom error page and/or a caching duration.
* `response_code` - (Optional) HTTP status code that you want CloudFront to return to the viewer along with the custom error page. Both `response_code` and `response_page_path` must be specified or both must be omitted.
* `response_page_path` - (Optional) Path to the custom error page that you want CloudFront to return to a viewer when your origin returns the HTTP status code specified by ErrorCode. Both `response_code` and `response_page_path` must be specified or both must be omitted.

### Tenant Config

* `parameter_definition` - (Required) One or more parameter definitions for the tenant configuration. See [Parameter Definition](#parameter-definition) below.

### Parameter Definition

* `name` - (Required) Name of the parameter.
* `definition` - (Required) Definition of the parameter schema. See [Parameter Definition Schema](#parameter-definition-schema) below.

### Parameter Definition Schema

* `string_schema` - (Required) String schema configuration. See [String Schema](#string-schema) below.

### String Schema

* `required` - (Required) Whether the parameter is required.
* `comment` - (Optional) Comment describing the parameter.
* `default_value` - (Optional) Default value for the parameter.

### Viewer Certificate

* `acm_certificate_arn` - (Optional) ARN of the AWS Certificate Manager certificate that you wish to use with this distribution. Required when using a custom SSL certificate.
* `cloudfront_default_certificate` - (Optional) Whether to use the CloudFront default certificate. Cannot be used with `acm_certificate_arn`.
* `minimum_protocol_version` - (Optional) Minimum version of the SSL protocol that you want CloudFront to use for HTTPS connections. Default: `TLSv1`.
* `ssl_support_method` - (Optional) How you want CloudFront to serve HTTPS requests. Valid values are `sni-only` and `vip`. Required when `acm_certificate_arn` is specified.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN for the distribution.
* `caller_reference` - Internal value used by CloudFront to allow future updates to the distribution configuration.
* `connection_mode` - Connection mode for the distribution. Always set to `tenant-only` for multi-tenant distributions.
* `domain_name` - Domain name corresponding to the distribution.
* `etag` - Current version of the distribution's information.
* `id` - Identifier for the distribution.
* `in_progress_invalidation_batches` - Number of invalidation batches currently in progress.
* `last_modified_time` - Date and time the distribution was last modified.
* `status` - Current status of the distribution. `Deployed` if the distribution's information is fully propagated throughout the Amazon CloudFront system.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `active_trusted_key_groups` - List of key groups that CloudFront can use to validate signed URLs or signed cookies. See [Active Trusted Key Groups](#active-trusted-key-groups) below.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Multi-tenant Distributions using the `id`. For example:

```terraform
import {
  to = aws_cloudfront_multitenant_distribution.distribution
  id = "E74FTE3AEXAMPLE"
}
```

Using `terraform import`, import CloudFront Multi-tenant Distributions using the `id`. For example:

```console
% terraform import aws_cloudfront_multitenant_distribution.distribution E74FTE3AEXAMPLE
```
