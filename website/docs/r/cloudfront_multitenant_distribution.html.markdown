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

For information about CloudFront multi-tenant distributions, see the [Amazon CloudFront Developer Guide][1].

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
- Cache behavior TTL settings (`default_ttl`, `max_ttl`, `min_ttl`) - Use cache policies instead
- `smooth_streaming` in cache behaviors
- `trusted_signers` in cache behaviors - Use `trusted_key_groups` instead
- `forwarded_values` in cache behaviors - Deprecated, use cache policies instead
- `aliases` - Managed by connection groups
- `is_ipv6_enabled` - Managed by connection groups
- `price_class` - Managed by connection groups
- `staging` mode
- `continuous_deployment_policy_id`
- `anycast_ip_list_id` - Use connection groups instead

## Example Usage

```terraform
resource "aws_cloudfront_multitenant_distribution" "example" {
  comment = "Multi-tenant distribution for my application"
  enabled = true

  origin {
    domain_name = "example.com"
    origin_id   = "example-origin"

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
    
    allowed_methods = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods  = ["GET", "HEAD"]
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

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `comment` - (Required) Any comments you want to include about the distribution.
* `enabled` - (Required) Whether the distribution is enabled to accept end user requests for content.
* `default_root_object` - (Optional) Object that you want CloudFront to return when an end user requests the root URL.
* `http_version` - (Optional) Maximum HTTP version to support on the distribution. Allowed values are `http1.1`, `http2`, `http2and3`, and `http3`. Default: `http2`.
* `web_acl_id` - (Optional) Unique identifier that specifies the AWS WAF v2 web ACL to associate with this distribution.
* `origin` - (Required) One or more origins for this distribution (multiples allowed). See [Origin](#origin) below.
* `default_cache_behavior` - (Required) Default cache behavior for this distribution. See [Default Cache Behavior](#default-cache-behavior) below.
* `cache_behavior` - (Optional) Ordered list of cache behaviors resource for this distribution. See [Cache Behavior](#cache-behavior) below.
* `custom_error_response` - (Optional) One or more custom error response elements. See [Custom Error Response](#custom-error-response) below.
* `logging_config` - (Optional) Logging configuration that controls how logs are written to your distribution. See [Logging Config](#logging-config) below.
* `origin_group` - (Optional) One or more origin_group for this distribution (multiples allowed). See [Origin Group](#origin-group) below.
* `restrictions` - (Required) Restriction configuration for this distribution. See [Restrictions](#restrictions) below.
* `viewer_certificate` - (Required) SSL configuration for this distribution. See [Viewer Certificate](#viewer-certificate) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

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
* `s3_origin_config` - (Optional) CloudFront S3 origin access identity to associate with the origin. See [S3 Origin Config](#s3-origin-config) below.

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

### Cache Behavior

Cache behavior supports all the same arguments as [Default Cache Behavior](#default-cache-behavior) with the addition of:

* `path_pattern` - (Required) Pattern that specifies which requests you want this cache behavior to apply to.

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

### Active Trusted Key Groups

* `enabled` - Whether any of the key groups have public keys that CloudFront can use to verify the signatures of signed URLs and signed cookies.
* `items` - List of key groups. See [Key Group Items](#key-group-items) below.

### Key Group Items

* `key_group_id` - ID of the key group that contains the public keys.
* `key_pair_ids` - Set of active CloudFront key pairs associated with the signer that can be used to verify the signatures of signed URLs and signed cookies.

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

[1]: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/
