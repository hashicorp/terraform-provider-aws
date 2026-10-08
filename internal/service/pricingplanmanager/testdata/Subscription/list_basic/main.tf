# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pricingplanmanager_subscription" "test" {
  count = var.resource_count

  plan_family = "CloudFront"
  plan_tier   = "FREE"

  resource_arns = [
    aws_cloudfront_distribution.test[count.index].arn,
    aws_wafv2_web_acl.test[count.index].arn,
  ]
}

resource "aws_cloudfront_distribution" "test" {
  count = var.resource_count

  enabled    = true
  comment    = "${var.rName}-${count.index}"
  web_acl_id = aws_wafv2_web_acl.test[count.index].arn

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    # Managed-CachingOptimized. Flat-rate plan eligibility requires modern
    # cache settings (a cache policy) rather than legacy forwarded_values.
    cache_policy_id = "658327ea-f89d-4fab-a63d-7e88639e58f6"
  }

  origin {
    domain_name = "www.example.com"
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}

resource "aws_wafv2_web_acl" "test" {
  # Web ACLs for CloudFront distributions must be created in us-east-1.
  region = "us-east-1"
  count  = var.resource_count

  name  = "${var.rName}-${count.index}"
  scope = "CLOUDFRONT"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = var.rName
    sampled_requests_enabled   = false
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
