# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_input_source" "test" {
  service_arn = aws_resiliencehubv2_service.test.arn

  resource_configuration {
    design_file_s3_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
  }
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "design.txt"
  source = "test-fixtures/design.txt"
}

resource "aws_s3_bucket" "test" {
  bucket = var.rName
}

resource "aws_resiliencehubv2_service" "test" {
  name    = var.rName
  regions = [data.aws_region.current.name]

  permission_model {
    invoker_role_name = aws_iam_role.test.name
  }

  depends_on = [aws_iam_role_policy_attachment.service_AWSResilienceHubV2AssessmentExecutionPolicy]
}

data "aws_region" "current" {
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "${var.rName}-invoker"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "resiliencehub.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "service_AWSResilienceHubV2AssessmentExecutionPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AWSResilienceHubV2AssessmentExecutionPolicy"
  role       = aws_iam_role.test.name
}

resource "aws_iam_policy" "s3_bucket_full_access" {
  name        = "${var.rName}-s3-bucket"
  description = "Full S3 permissions for one bucket and its objects"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "s3:*"
      Resource = [
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.bucket}",
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.bucket}/*",
      ]
    }]
  })
}

resource "aws_iam_role_policy_attachment" "s3_bucket_full_access" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.s3_bucket_full_access.arn
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
