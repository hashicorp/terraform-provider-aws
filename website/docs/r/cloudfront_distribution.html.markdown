---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_distribution"
description: |-
  Provides a CloudFront web distribution resource.
---

# Resource: aws_cloudfront_distribution

Creates an Amazon CloudFront web distribution.

For information about CloudFront distributions, see the [Amazon CloudFront Developer Guide][1]. For specific information about creating CloudFront web distributions, see the [POST Distribution][2] page in the Amazon CloudFront API Reference.

~> **NOTE:** CloudFront distributions take about 15 minutes to reach a deployed state after creation or modification. During this time, deletes to resources will be blocked. If you need to delete a distribution that is enabled and you do not want to wait, you need to use the `retain_on_delete` flag.

## Example Usage

### S3 Origin

The example below creates a CloudFront distribution with an S3 origin.

```terraform
resource "aws_s3_bucket" "b" {
  bucket = "mybucket"

  tags = {
    Name = "My bucket"
  }
}

# See https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html
data "aws_iam_policy_document" "origin_bucket_policy" {
  statement {
    sid    = "AllowCloudFrontServicePrincipalReadWrite"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["cloudfront.amazonaws.com"]
    }

    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]

    resources = [
      "${aws_s3_bucket.b.arn}/*",
    ]

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceArn"
      values   = [aws_cloudfront_distribution.s3_distribution.arn]
    }
  }
}

resource "aws_s3_bucket_policy" "b" {
  bucket = aws_s3_bucket.b.bucket
  policy = data.aws_iam_policy_document.origin_bucket_policy.json
}

locals {
  s3_origin_id = "myS3Origin"
  my_domain    = "mydomain.com"
}

data "aws_acm_certificate" "my_domain" {
  region   = "us-east-1"
  domain   = "*.${local.my_domain}"
  statuses = ["ISSUED"]
}

resource "aws_cloudfront_origin_access_control" "default" {
  name                              = "default-oac"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

resource "aws_cloudfront_distribution" "s3_distribution" {
  origin {
    domain_name              = aws_s3_bucket.b.bucket_regional_domain_name
    origin_access_control_id = aws_cloudfront_origin_access_control.default.id
    origin_id                = local.s3_origin_id
  }

  enabled             = true
  is_ipv6_enabled     = true
  comment             = "Some comment"
  default_root_object = "index.html"

  aliases = ["mysite.${local.my_domain}", "yoursite.${local.my_domain}"]

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = local.s3_origin_id

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  # Cache behavior with precedence 0
  ordered_cache_behavior {
    path_pattern     = "/content/immutable/*"
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD", "OPTIONS"]
    target_origin_id = local.s3_origin_id

    forwarded_values {
      query_string = false
      headers      = ["Origin"]

      cookies {
        forward = "none"
      }
    }

    min_ttl                = 0
    default_ttl            = 86400
    max_ttl                = 31536000
    compress               = true
    viewer_protocol_policy = "redirect-to-https"
  }

  # Cache behavior with precedence 1
  ordered_cache_behavior {
    path_pattern     = "/content/*"
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = local.s3_origin_id

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
    compress               = true
    viewer_protocol_policy = "redirect-to-https"
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  tags = {
    Environment = "production"
  }

  viewer_certificate {
    acm_certificate_arn = data.aws_acm_certificate.my_domain.arn
    ssl_support_method  = "sni-only"
  }
}

# Create Route53 records for the CloudFront distribution aliases
data "aws_route53_zone" "my_domain" {
  name = local.my_domain
}

resource "aws_route53_record" "cloudfront" {
  for_each = aws_cloudfront_distribution.s3_distribution.aliases
  zone_id  = data.aws_route53_zone.my_domain.zone_id
  name     = each.value
  type     = "A"

  alias {
    name                   = aws_cloudfront_distribution.s3_distribution.domain_name
    zone_id                = aws_cloudfront_distribution.s3_distribution.hosted_zone_id
    evaluate_target_health = false
  }
}
```

### With Failover Routing

The example below creates a CloudFront distribution with an origin group for failover routing.

```terraform
resource "aws_cloudfront_distribution" "s3_distribution" {
  origin_group {
    origin_id = "groupS3"

    failover_criteria {
      status_codes = [403, 404, 500, 502]
    }

    member {
      origin_id = "primaryS3"
    }

    member {
      origin_id = "failoverS3"
    }
  }

  origin {
    domain_name = aws_s3_bucket.primary.bucket_regional_domain_name
    origin_id   = "primaryS3"

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.default.cloudfront_access_identity_path
    }
  }

  origin {
    domain_name = aws_s3_bucket.failover.bucket_regional_domain_name
    origin_id   = "failoverS3"

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.default.cloudfront_access_identity_path
    }
  }

  default_cache_behavior {
    # ... other configuration ...
    target_origin_id = "groupS3"
  }

  # ... other configuration ...
}
```

### With Managed Caching Policy

The example below creates a CloudFront distribution with an [AWS managed caching policy](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/using-managed-cache-policies.html).

```terraform
locals {
  s3_origin_id = "myS3Origin"
}

resource "aws_cloudfront_distribution" "s3_distribution" {
  origin {
    domain_name = aws_s3_bucket.primary.bucket_regional_domain_name
    origin_id   = "myS3Origin"

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.default.cloudfront_access_identity_path
    }
  }
  enabled             = true
  is_ipv6_enabled     = true
  comment             = "Some comment"
  default_root_object = "index.html"

  # AWS Managed Caching Policy (CachingDisabled)
  default_cache_behavior {
    # Using the CachingDisabled managed policy ID:
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad"
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = local.s3_origin_id
    viewer_protocol_policy = "allow-all"
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  # ... other configuration ...
}
```

### With V2 logging to S3

The example below creates a CloudFront distribution with [standard logging V2 to S3](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/standard-logging.html#enable-access-logging-api).

```terraform
resource "aws_cloudfront_distribution" "example" {
  # other config...
}

resource "aws_cloudwatch_log_delivery_source" "example" {
  region = "us-east-1"

  name         = "example"
  log_type     = "ACCESS_LOGS"
  resource_arn = aws_cloudfront_distribution.example.arn
}

resource "aws_s3_bucket" "example" {
  bucket        = "testbucket"
  force_destroy = true
}

resource "aws_cloudwatch_log_delivery_destination" "example" {
  region = "us-east-1"

  name          = "s3-destination"
  output_format = "parquet"

  delivery_destination_configuration {
    destination_resource_arn = "${aws_s3_bucket.example.arn}/prefix"
  }
}

resource "aws_cloudwatch_log_delivery" "example" {
  region = "us-east-1"

  delivery_source_name     = aws_cloudwatch_log_delivery_source.example.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.example.arn

  s3_delivery_configuration {
    suffix_path = "/123456678910/{DistributionId}/{yyyy}/{MM}/{dd}/{HH}"
  }
}
```

### With V2 logging to Data Firehose

The example below creates a CloudFront distribution with [standard logging V2 to Data Firehose](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/standard-logging.html#enable-access-logging-api).

```terraform
resource "aws_cloudfront_distribution" "example" {
  # other config
}

resource "aws_kinesis_firehose_delivery_stream" "cloudfront_logs" {
  region = "us-east-1"
  # The tag named "LogDeliveryEnabled" must be set to "true" to allow the service-linked role "AWSServiceRoleForLogDelivery"
  # to perform permitted actions on your behalf.
  # See: https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/AWS-logs-and-resource-policy.html#AWS-logs-infrastructure-Firehose
  tags = {
    LogDeliveryEnabled = "true"
  }

  # other config
}

resource "aws_cloudwatch_log_delivery_source" "example" {
  region = "us-east-1"

  name         = "cloudfront-logs-source"
  log_type     = "ACCESS_LOGS"
  resource_arn = aws_cloudfront_distribution.example.arn
}

resource "aws_cloudwatch_log_delivery_destination" "example" {
  region = "us-east-1"

  name          = "firehose-destination"
  output_format = "json"
  delivery_destination_configuration {
    destination_resource_arn = aws_kinesis_firehose_delivery_stream.cloudfront_logs.arn
  }
}
resource "aws_cloudwatch_log_delivery" "example" {
  region = "us-east-1"

  delivery_source_name     = aws_cloudwatch_log_delivery_source.example.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `aliases` (Optional) - Extra CNAMEs (alternate domain names), if any, for this distribution.
* `anycast_ip_list_id` (Optional) - ID of the Anycast static IP list that is associated with the distribution.
* `comment` (Optional) - Any comments you want to include about the distribution.
* `continuous_deployment_policy_id` (Optional) - Identifier of a continuous deployment policy. This argument should only be set on a production distribution. See the [`aws_cloudfront_continuous_deployment_policy` resource](./cloudfront_continuous_deployment_policy.html.markdown) for additional details.
* `custom_error_response` (Optional) - One or more [custom error response](#custom-error-response-arguments) elements (multiples allowed).
* `default_cache_behavior` (Required) - [Default cache behavior](#default-cache-behavior-arguments) for this distribution (maximum one). Requires either `cache_policy_id` (preferred) or `forwarded_values` (deprecated) be set.
* `default_root_object` (Optional) - Object that you want CloudFront to return (for example, index.html) when an end user requests the root URL.
* `enabled` (Required) - Whether the distribution is enabled to accept end user requests for content.
* `is_ipv6_enabled` (Optional) - Whether the IPv6 is enabled for the distribution.
* `http_version` (Optional) - Maximum HTTP version to support on the distribution. Allowed values are `http1.1`, `http2`, `http2and3` and `http3`. The default is `http2`.
* `logging_config` (Optional) - The [logging configuration](#logging-config-arguments) that controls how logs are written to your distribution (maximum one). AWS provides two versions of access logs for CloudFront: Legacy and v2. This argument configures legacy version standard logs.
* `ordered_cache_behavior` (Optional) - Ordered list of [cache behaviors](#cache-behavior-arguments) resource for this distribution. List from top to bottom in order of precedence. The topmost cache behavior will have precedence 0.
* `origin` (Required) - One or more [origins](#origin-arguments) for this distribution (multiples allowed).
* `origin_group` (Optional) - One or more [origin_group](#origin-group-arguments) for this distribution (multiples allowed).
* `price_class` (Optional) - Price class for this distribution. One of `PriceClass_All`, `PriceClass_200`, `PriceClass_100`.
* `restrictions` (Required) - The [restriction configuration](#restrictions-arguments) for this distribution (maximum one).
* `staging` (Optional) - A Boolean that indicates whether this is a staging distribution. Defaults to `false`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `viewer_certificate` (Required) - The [SSL configuration](#viewer-certificate-arguments) for this distribution (maximum one).
* `web_acl_id` (Optional) - Unique identifier that specifies the AWS WAF web ACL, if any, to associate with this distribution. To specify a web ACL created using the latest version of AWS WAF (WAFv2), use the ACL ARN, for example `aws_wafv2_web_acl.example.arn`. To specify a web ACL created using AWS WAF Classic, use the ACL ID, for example `aws_waf_web_acl.example.id`. The WAF Web ACL must exist in the WAF Global (CloudFront) region and the credentials configuring this argument must have `waf:GetWebACL` permissions assigned.
* `retain_on_delete` (Optional) - Disables the distribution instead of deleting it when destroying the resource through Terraform. If this is set, the distribution needs to be deleted manually afterwards. Default: `false`.
* `wait_for_deployment` (Optional) - If enabled, the resource will wait for the distribution status to change from `InProgress` to `Deployed`. Setting this to`false` will skip the process. Default: `true`.

#### Cache Behavior Arguments

~> **NOTE:** To achieve the setting of 'Use origin cache headers' without a linked cache policy, use the following TTL values: `min_ttl` = 0, `max_ttl` = 31536000, `default_ttl` = 86400. See [this issue](https://github.com/hashicorp/terraform-provider-aws/issues/19382) for additional context.

* `allowed_methods` (Required) - Controls which HTTP methods CloudFront processes and forwards to your Amazon S3 bucket or your custom origin.
* `cached_methods` (Required) - Controls whether CloudFront caches the response to requests using the specified HTTP methods.
* `cache_policy_id` (Optional) - Unique identifier of the cache policy that is attached to the cache behavior. If configuring the `default_cache_behavior` either `cache_policy_id` or `forwarded_values` must be set.
* `compress` (Optional) - Whether you want CloudFront to automatically compress content for web requests that include `Accept-Encoding: gzip` in the request header (default: `false`).
* `default_ttl` (Optional) - Default amount of time (in seconds) that an object is in a CloudFront cache before CloudFront forwards another request in the absence of an `Cache-Control max-age` or `Expires` header. The TTL defined in Cache Policy overrides this configuration.
* `field_level_encryption_id` (Optional) - Field level encryption configuration ID.
* `forwarded_values` (Optional, **Deprecated** use `cache_policy_id` or `origin_request_policy_id ` instead) - The [forwarded values configuration](#forwarded-values-arguments) that specifies how CloudFront handles query strings, cookies and headers (maximum one).
* `lambda_function_association` (Optional) - A [config block](#lambda-function-association) that triggers a lambda function with specific actions (maximum 4).
* `function_association` (Optional) - A [config block](#function-association) that triggers a cloudfront function with specific actions (maximum 2).
* `max_ttl` (Optional) - Maximum amount of time (in seconds) that an object is in a CloudFront cache before CloudFront forwards another request to your origin to determine whether the object has been updated. Only effective in the presence of `Cache-Control max-age`, `Cache-Control s-maxage`, and `Expires` headers. The TTL defined in Cache Policy overrides this configuration.
* `min_ttl` (Optional) - Minimum amount of time that you want objects to stay in CloudFront caches before CloudFront queries your origin to see whether the object has been updated. Defaults to 0 seconds. The TTL defined in Cache Policy overrides this configuration.
* `origin_request_policy_id` (Optional) - Unique identifier of the origin request policy that is attached to the behavior.
* `path_pattern` (Required) - Pattern (for example, `images/*.jpg`) that specifies which requests you want this cache behavior to apply to.
* `realtime_log_config_arn` (Optional) - ARN of the [real-time log configuration](cloudfront_realtime_log_config.html) that is attached to this cache behavior.
* `response_headers_policy_id` (Optional) - Identifier for a response headers policy.
* `smooth_streaming` (Optional) - Indicates whether you want to distribute media files in Microsoft Smooth Streaming format using the origin that is associated with this cache behavior.
* `target_origin_id` (Required) - Value of ID for the origin that you want CloudFront to route requests to when a request matches the path pattern either for a cache behavior or for the default cache behavior.
* `trusted_key_groups` (Optional) - List of key group IDs that CloudFront can use to validate signed URLs or signed cookies. See the [CloudFront User Guide](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-trusted-signers.html) for more information about this feature.
* `trusted_signers` (Optional) - List of AWS account IDs (or `self`) that you want to allow to create signed URLs for private content. See the [CloudFront User Guide](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-trusted-signers.html) for more information about this feature.
* `viewer_protocol_policy` (Required) - Use this element to specify the protocol that users can use to access the files in the origin specified by TargetOriginId when a request matches the path pattern in PathPattern. One of `allow-all`, `https-only`, or `redirect-to-https`.
* `grpc_config` (Optional) - A [config block](#grpc-config-arguments) that sets the grpc config.

##### Forwarded Values Arguments

* `cookies` (Required) - The [forwarded values cookies](#cookies-arguments) that specifies how CloudFront handles cookies (maximum one).
* `headers` (Optional) - Headers, if any, that you want CloudFront to vary upon for this cache behavior. Specify `*` to include all headers.
* `query_string` (Required) - Indicates whether you want CloudFront to forward query strings to the origin that is associated with this cache behavior.
* `query_string_cache_keys` (Optional) - When specified, along with a value of `true` for `query_string`, all query strings are forwarded, however only the query string keys listed in this argument are cached. When omitted with a value of `true` for `query_string`, all query string keys are cached.

##### Lambda Function Association

Lambda@Edge allows you to associate an AWS Lambda Function with a predefined
event. You can associate a single function per event type. See [What is
Lambda@Edge](http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/what-is-lambda-at-edge.html)
for more information.

Example configuration:

```terraform
resource "aws_cloudfront_distribution" "example" {
  # ... other configuration ...

  # lambda_function_association is also supported by default_cache_behavior
  ordered_cache_behavior {
    # ... other configuration ...

    lambda_function_association {
      event_type   = "viewer-request"
      lambda_arn   = aws_lambda_function.example.qualified_arn
      include_body = false
    }
  }
}
```

* `event_type` (Required) - Specific event to trigger this function. Valid values: `viewer-request`, `origin-request`, `viewer-response`, `origin-response`.
* `lambda_arn` (Required) - ARN of the Lambda function.
* `include_body` (Optional) - When set to true it exposes the request body to the lambda function. Defaults to false. Valid values: `true`, `false`.

##### Function Association

With CloudFront Functions in Amazon CloudFront, you can write lightweight functions in JavaScript for high-scale, latency-sensitive CDN customizations. You can associate a single function per event type. See [CloudFront Functions](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/cloudfront-functions.html)
for more information.

Example configuration:

```terraform
resource "aws_cloudfront_distribution" "example" {
  # ... other configuration ...

  # function_association is also supported by default_cache_behavior
  ordered_cache_behavior {
    # ... other configuration ...

    function_association {
      event_type   = "viewer-request"
      function_arn = aws_cloudfront_function.example.arn
    }
  }
}
```

* `event_type` (Required) - Specific event to trigger this function. Valid values: `viewer-request` or `viewer-response`.
* `function_arn` (Required) - ARN of the CloudFront function.

##### GRPC config arguments

* `enabled` (Required) - Whether Grpc requests are enabled.

##### Cookies Arguments

* `forward` (Required) - Whether you want CloudFront to forward cookies to the origin that is associated with this cache behavior. You can specify `all`, `none` or `whitelist`. If `whitelist`, you must include the subsequent `whitelisted_names`.
* `whitelisted_names` (Optional) - If you have specified `whitelist` to `forward`, the whitelisted cookies that you want CloudFront to forward to your origin.

#### Custom Error Response Arguments

~> **NOTE:** When specifying either `response_page_path` or `response_code`, **both** must be set.

* `error_caching_min_ttl` (Optional) - Minimum amount of time you want HTTP error codes to stay in CloudFront caches before CloudFront queries your origin to see whether the object has been updated.
* `error_code` (Required) - 4xx or 5xx HTTP status code that you want to customize.
* `response_code` (Optional) - HTTP status code that you want CloudFront to return with the custom error page to the viewer.
* `response_page_path` (Optional) - Path of the custom error page (for example, `/custom_404.html`).

#### Default Cache Behavior Arguments

The arguments for `default_cache_behavior` are the same as for
[`ordered_cache_behavior`](#cache-behavior-arguments), except for the `path_pattern`
argument should not be specified.

#### Logging Config Arguments

* `bucket` (Optional) - Amazon S3 bucket for V1 logging where access logs are stored, for example, `myawslogbucket.s3.amazonaws.com`. V1 logging is enabled when this argument is specified. The bucket must have correct ACL attached with "FULL_CONTROL" permission for "awslogsdelivery" account (Canonical ID: "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0") for log transfer to work.
* `include_cookies` (Optional) - Whether to include cookies in access logs (default: `false`). This argument applies to both V1 and V2 logging.
* `prefix` (Optional) - Prefix added to the access log file names for V1 logging, for example, `myprefix/`. This argument is effective only when V1 logging is enabled.

#### Origin Arguments

* `connection_attempts` (Optional) - Number of times that CloudFront attempts to connect to the origin. Must be between 1-3. Defaults to 3.
* `connection_timeout` (Optional) - Number of seconds that CloudFront waits when trying to establish a connection to the origin. Must be between 1-10. Defaults to 10.
* `custom_origin_config` - The [CloudFront custom origin](#custom-origin-config-arguments) configuration information. If an S3 origin is required, use `origin_access_control_id` or `s3_origin_config` instead.
* `domain_name` (Required) - DNS domain name of either the S3 bucket, or web site of your custom origin.
* `custom_header` (Optional) - One or more sub-resources with `name` and `value` parameters that specify header data that will be sent to the origin (multiples allowed).
* `origin_access_control_id` (Optional) - Unique identifier of a [CloudFront origin access control][8] for this origin.
* `origin_id` (Required) - Unique identifier for the origin.
* `origin_path` (Optional) - Optional element that causes CloudFront to request your content from a directory in your Amazon S3 bucket or your custom origin.
* `origin_shield` - (Optional) [CloudFront Origin Shield](#origin-shield-arguments) configuration information. Using Origin Shield can help reduce the load on your origin. For more information, see [Using Origin Shield](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/origin-shield.html) in the Amazon CloudFront Developer Guide.
* `response_completion_timeout` - (Optional) Time (in seconds) that a request from CloudFront to the origin can stay open and wait for a response. Must be integer greater than or equal to the value of `origin_read_timeout`. If omitted or explicitly set to `0`, no maximum value is enforced.
* `s3_origin_config` - (Optional) [CloudFront S3 origin](#s3-origin-config-arguments) configuration information. If a custom origin is required, use `custom_origin_config` instead.
* `vpc_origin_config` - (Optional) The [VPC origin configuration](#vpc-origin-config-arguments).

##### Custom Origin Config Arguments

* `http_port` (Required) - HTTP port the custom origin listens on.
* `https_port` (Required) - HTTPS port the custom origin listens on.
* `ip_address_type` (Optional) - IP protocol CloudFront uses when connecting to your origin. Valid values: `ipv4`, `ipv6`, `dualstack`.
* `origin_protocol_policy` (Required) - Origin protocol policy to apply to your origin. One of `http-only`, `https-only`, or `match-viewer`.
* `origin_ssl_protocols` (Required) - List of SSL/TLS protocols that CloudFront can use when connecting to your origin over HTTPS. Valid values: `SSLv3`, `TLSv1`, `TLSv1.1`, `TLSv1.2`. For more information, see [Minimum Origin SSL Protocol](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/distribution-web-values-specify.html#DownloadDistValuesOriginSSLProtocols) in the Amazon CloudFront Developer Guide.
* `origin_keepalive_timeout` - (Optional) The Custom KeepAlive timeout, in seconds. By default, AWS enforces an upper limit of `60`. But you can request an [increase](http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/RequestAndResponseBehaviorCustomOrigin.html#request-custom-request-timeout). Defaults to `5`.
* `origin_read_timeout` - (Optional) The Custom Read timeout, in seconds. By default, AWS enforces an upper limit of `60`. But you can request an [increase](http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/RequestAndResponseBehaviorCustomOrigin.html#request-custom-request-timeout). Defaults to `30`.

##### Origin Shield Arguments

* `enabled` (Required) - Whether Origin Shield is enabled.
* `origin_shield_region` (Optional) - AWS Region for Origin Shield. To specify a region, use the region code, not the region name. For example, specify the US East (Ohio) region as `us-east-2`.

##### S3 Origin Config Arguments

* `origin_access_identity` (Required) - The [CloudFront origin access identity][5] to associate with the origin.

##### VPC Origin Config Arguments

* `origin_keepalive_timeout` - (Optional) Specifies how long, in seconds, CloudFront persists its connection to the origin. The minimum timeout is 1 second, the maximum is 60 seconds. Defaults to `5`.
* `origin_read_timeout` - (Optional) Specifies how long, in seconds, CloudFront waits for a response from the origin. This is also known as the _origin response timeout_. The minimum timeout is 1 second, the maximum is 60 seconds. Defaults to `30`.
* `owner_account_id` - (Optional) The AWS account ID that owns the VPC origin. Required when referencing a VPC origin from a different AWS account for cross-account VPC origin access.
* `vpc_origin_id` (Required) - The VPC origin ID.

#### Origin Group Arguments

* `origin_id` (Required) - Unique identifier for the origin group.
* `failover_criteria` (Required) - The [failover criteria](#failover-criteria-arguments) for when to failover to the secondary origin.
* `member` (Required) - Ordered [member](#member-arguments) configuration blocks assigned to the origin group, where the first member is the primary origin. You must specify two members.

##### Failover Criteria Arguments

* `status_codes` (Required) - List of HTTP status codes for the origin group.

##### Member Arguments

* `origin_id` (Required) - Unique identifier of the member origin.

#### Restrictions Arguments

The `restrictions` sub-resource takes another single sub-resource named `geo_restriction` (see the example for usage).

The arguments of `geo_restriction` are:

* `locations` (Required) - [ISO 3166-1-alpha-2 codes][4] for which you want CloudFront either to distribute your content (`whitelist`) or not distribute your content (`blacklist`). If the type is specified as `none` an empty array can be used.
* `restriction_type` (Required) - Method that you want to use to restrict distribution of your content by country: `none`, `whitelist`, or `blacklist`.

#### Viewer Certificate Arguments

* `acm_certificate_arn` - ARN of the [AWS Certificate Manager][6] certificate that you wish to use with this distribution. Specify this, `cloudfront_default_certificate`, or `iam_certificate_id`.  The ACM certificate must be in  US-EAST-1.
* `cloudfront_default_certificate` - `true` if you want viewers to use HTTPS to request your objects and you're using the CloudFront domain name for your distribution. Specify this, `acm_certificate_arn`, or `iam_certificate_id`.
* `iam_certificate_id` - IAM certificate identifier of the custom viewer certificate for this distribution if you are using a custom domain. Specify this, `acm_certificate_arn`, or `cloudfront_default_certificate`.
* `minimum_protocol_version` - Minimum version of the SSL protocol that you want CloudFront to use for HTTPS connections. Can only be set if `cloudfront_default_certificate = false`. See all possible values in [this](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/secure-connections-supported-viewer-protocols-ciphers.html) table under "Security policy." Some examples include: `TLSv1.2_2019` and `TLSv1.2_2021`. Default: `TLSv1`. **NOTE**: If you are using a custom certificate (specified with `acm_certificate_arn` or `iam_certificate_id`), and have specified `sni-only` in `ssl_support_method`, `TLSv1` or later must be specified. If you have specified `vip` in `ssl_support_method`, only `SSLv3` or `TLSv1` can be specified. If you have specified `cloudfront_default_certificate`, `TLSv1` must be specified.
* `ssl_support_method` - How you want CloudFront to serve HTTPS requests. One of `vip`, `sni-only`, or `static-ip`. Required if you specify `acm_certificate_arn` or `iam_certificate_id`. **NOTE:** `vip` causes CloudFront to use a dedicated IP address and may incur extra charges.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier for the distribution. For example: `EDFDVBD632BHDS5`.
* `arn` - ARN for the distribution. For example: `arn:aws:cloudfront::123456789012:distribution/EDFDVBD632BHDS5`, where `123456789012` is your AWS account ID.
* `caller_reference` - Internal value used by CloudFront to allow future updates to the distribution configuration.
* `logging_v1_enabled` - Whether V1 logging is enabled for the distribution.
* `status` - Current status of the distribution. `Deployed` if the distribution's information is fully propagated throughout the Amazon CloudFront system.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `trusted_key_groups` - List of nested attributes for active trusted key groups, if the distribution is set up to serve private content with signed URLs.
    * `enabled` - `true` if any of the key groups have public keys that CloudFront can use to verify the signatures of signed URLs and signed cookies.
    * `items` - List of nested attributes for each key group.
        * `key_group_id` - ID of the key group that contains the public keys.
        * `key_pair_ids` - Set of CloudFront key pair IDs.
* `trusted_signers` - List of nested attributes for active trusted signers, if the distribution is set up to serve private content with signed URLs.
    * `enabled` - `true` if any of the AWS accounts listed as trusted signers have active CloudFront key pairs
    * `items` - List of nested attributes for each trusted signer
        * `aws_account_number` - AWS account ID or `self`
        * `key_pair_ids` - Set of active CloudFront key pairs associated with the signer account
* `domain_name` - Domain name corresponding to the distribution. For example: `d604721fxaaqy9.cloudfront.net`.
* `last_modified_time` - Date and time the distribution was last modified.
* `in_progress_validation_batches` - Number of invalidation batches currently in progress.
* `etag` - Current version of the distribution's information. For example: `E2QWRUHAPOMQZL`.
* `hosted_zone_id` - CloudFront Route 53 zone ID that can be used to route an [Alias Resource Record Set][7] to. This attribute is simply an alias for the zone ID `Z2FDTNDATAQYW2`.

[1]: http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/Introduction.html
[2]: https://docs.aws.amazon.com/cloudfront/latest/APIReference/API_CreateDistribution.html
[4]: http://www.iso.org/iso/country_codes/iso_3166_code_lists/country_names_and_code_elements.htm
[5]: /docs/providers/aws/r/cloudfront_origin_access_identity.html
[6]: https://aws.amazon.com/certificate-manager/
[7]: http://docs.aws.amazon.com/Route53/latest/APIReference/CreateAliasRRSAPI.html
[8]: /docs/providers/aws/r/cloudfront_origin_access_control.html

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Distributions using the `id`. For example:

```terraform
import {
  to = aws_cloudfront_distribution.distribution
  id = "E74FTE3EXAMPLE"
}
```

Using `terraform import`, import CloudFront Distributions using the `id`. For example:

```console
% terraform import aws_cloudfront_distribution.distribution E74FTE3EXAMPLE
```
