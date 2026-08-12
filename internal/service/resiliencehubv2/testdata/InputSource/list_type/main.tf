# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_input_source" "test" {
  count = 2

  service_arn = aws_resiliencehubv2_service.test.arn

  resource_configuration {
    cfn_stack_arn     = count.index == 0 ? aws_cloudformation_stack.test.id : null
    tf_state_file_url = count.index == 1 ? "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}" : null
  }
}

resource "aws_cloudformation_stack" "test" {
  name = var.rName

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Description              = "Test stack for NGRH input source"
    Resources = {
      WaitHandle = {
        Type = "AWS::CloudFormation::WaitConditionHandle"
      }
    }
  })
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "tf_state"
  source = "test-fixtures/terraform.tfstate.json"
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
  name        = "s3-bucket-full-access"
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
