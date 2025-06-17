# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrock_model_invocation_logging_configuration" "test" {
  region = var.region

  depends_on = [
    aws_s3_bucket_policy.test,
    aws_iam_role_policy_attachment.test,
  ]

  logging_config {
    cloudwatch_config {
      log_group_name = aws_cloudwatch_log_group.test.name
      role_arn       = aws_iam_role.test.arn
    }

    s3_config {
      bucket_name = aws_s3_bucket.test.bucket
      key_prefix  = "bedrock"
    }
  }
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {
  region = var.region
}
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  region = var.region

  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  region = var.region

  bucket = aws_s3_bucket.test.bucket

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "bedrock.amazonaws.com"
    },
    "Action": [
      "s3:*"
    ],
    "Resource": [
      "${aws_s3_bucket.test.arn}/*"
    ],
    "Condition": {
      "StringEquals": {
        "aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
      },
      "ArnLike": {
        "aws:SourceArn": "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      }
    }
  }]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  region = var.region

  name = var.rName
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "bedrock.amazonaws.com"
    },
    "Action": "sts:AssumeRole",
    "Condition": {
      "StringEquals": {
        "aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
      },
      "ArnLike": {
        "aws:SourceArn": "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      }
    }
  }]
}  
EOF
}

resource "aws_iam_policy" "test" {
  name        = var.rName
  path        = "/"
  description = "BedrockCloudWatchPolicy"

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [{
      "Effect" : "Allow",
      "Action" : [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource" : "${aws_cloudwatch_log_group.test.arn}:log-stream:aws/bedrock/modelinvocations"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
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
