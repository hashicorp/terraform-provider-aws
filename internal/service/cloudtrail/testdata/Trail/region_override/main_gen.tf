# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudtrail" "test" {
  region = var.region

  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = var.rName
  s3_bucket_name = aws_s3_bucket.test.bucket
}

# testAccCloudTrailConfig_base

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {
  region = var.region

}

resource "aws_s3_bucket" "test" {
  region = var.region

  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  region = var.region

  bucket = aws_s3_bucket.test.bucket
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AWSCloudTrailAclCheck"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:GetBucketAcl"
        Resource = aws_s3_bucket.test.arn
        Condition = {
          StringEquals = {
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:cloudtrail:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:trail/${var.rName}"
          }
        }
      },
      {
        Sid    = "AWSCloudTrailWrite"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.test.arn}/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl"  = "bucket-owner-full-control"
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:cloudtrail:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:trail/${var.rName}"
          }
        }
      }
    ]
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
