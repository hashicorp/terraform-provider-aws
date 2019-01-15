provider "aws" {
  region = "${var.aws_region}"
}

resource "aws_s3_bucket" "primary" {
  bucket = "${var.primary_bucket_name}"
  acl    = "private"
}

resource "aws_s3_bucket" "backup" {
  bucket = "${var.backup_bucket_name}"
  acl    = "private"
}

locals {
  primary_s3_origin = "primaryS3"
  backup_s3_origin  = "failoverS3"
  origin_group      = "groupS3"
}

resource "aws_cloudfront_origin_access_identity" "default" {
  comment = "Testing Origin Access Identity"
}

resource "aws_cloudfront_distribution" "failover_distribution" {
  comment = "testing distribution for origin failover"
  enabled = true

  restrictions = {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  origin {
    domain_name = "${aws_s3_bucket.primary.bucket_regional_domain_name}"
    origin_id   = "${local.primary_s3_origin}"

    s3_origin_config {
      origin_access_identity = "${aws_cloudfront_origin_access_identity.default.cloudfront_access_identity_path}"
    }
  }

  origin {
    domain_name = "${aws_s3_bucket.backup.bucket_regional_domain_name}"
    origin_id   = "${local.backup_s3_origin}"

    s3_origin_config {
      origin_access_identity = "${aws_cloudfront_origin_access_identity.default.cloudfront_access_identity_path}"
    }
  }

  origin_group {
    origin_id = "${local.origin_group}"

    failover_criteria {
      status_codes = [403, 404, 500, 502, 503, 504]
    }

    members {
      ordered_origin_group_member {
        origin_id = "${local.primary_s3_origin}"
      }

      ordered_origin_group_member {
        origin_id = "${local.backup_s3_origin}"
      }
    }
  }

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "${local.origin_group}"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
