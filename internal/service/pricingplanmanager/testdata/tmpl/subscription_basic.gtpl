resource "aws_pricingplanmanager_subscription" "test" {
  plan_family = "CloudFront"
  plan_tier   = "FREE"

  resource_arns = [
    aws_cloudfront_distribution.test.arn,
    aws_wafv2_web_acl.test.arn,
  ]
}

resource "aws_cloudfront_distribution" "test" {
  enabled    = false
  comment    = var.rName
  web_acl_id = aws_wafv2_web_acl.test.arn

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
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

  name  = var.rName
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
