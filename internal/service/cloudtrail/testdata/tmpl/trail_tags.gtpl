resource "aws_cloudtrail" "test" {
{{- template "region" }}
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = var.rName
  s3_bucket_name = aws_s3_bucket.test.bucket
{{- template "tags" . }}
}

# testAccCloudTrailConfig_base

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {
{{- template "region" }}
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
{{- template "region" }}
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
