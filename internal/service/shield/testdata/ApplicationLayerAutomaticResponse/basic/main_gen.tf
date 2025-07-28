# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_shield_application_layer_automatic_response" "test" {
  resource_arn = aws_cloudfront_distribution.test.arn
  action       = "COUNT"

  depends_on = [
    aws_shield_protection.test,
    aws_cloudfront_distribution.test,
    aws_wafv2_web_acl.test
  ]
}

resource "aws_shield_protection" "test" {
  name         = var.rName
  resource_arn = aws_cloudfront_distribution.test.arn
}

resource "aws_wafv2_web_acl" "test" {
  name        = var.rName
  description = var.rName
  scope       = "CLOUDFRONT"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = var.rName
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [
      rule,
    ]
  }
}

resource "aws_cloudfront_distribution" "test" {
  origin {
    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols = [
        "TLSv1",
        "TLSv1.1",
        "TLSv1.2",
      ]
    }

    domain_name = "${var.rName}.com"
    origin_id   = var.rName
  }

  enabled             = false
  wait_for_deployment = false
  web_acl_id          = aws_wafv2_web_acl.test.arn
  default_cache_behavior {
    allowed_methods  = ["HEAD", "DELETE", "POST", "GET", "OPTIONS", "PUT", "PATCH"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = var.rName
    forwarded_values {
      query_string = false
      headers      = ["*"]
      cookies {
        forward = "none"
      }
    }
    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 0
    max_ttl                = 0
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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
