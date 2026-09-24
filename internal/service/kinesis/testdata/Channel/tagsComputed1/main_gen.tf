# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_kinesis_channel" "test" {
  channel_name               = var.rName
  service_execution_role_arn = aws_iam_role.role.arn

  stream_configuration_list {
    stream_arn = aws_kinesis_stream.stream.arn

    record_configuration {
      record_format_type = "JSON"
    }
  }

  s3_destination_configuration {
    storage_configuration {
      bucket_arn            = aws_s3_bucket.bucket.arn
      expected_bucket_owner = data.aws_caller_identity.current.account_id
      compression_type      = "NONE"
    }
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }

  depends_on = [
    aws_iam_role_policy.policy,
  ]
}

resource "aws_iam_role" "role" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = [
            "kinesis.amazonaws.com",
            "s3.amazonaws.com"
          ]
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy" "policy" {
  name = var.rName
  role = aws_iam_role.role.id

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [
      {
        Effect = "Allow"
        Action = [
          "kinesis:*"
        ]
        Resource = aws_kinesis_stream.stream.arn
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:DescribeLogStreams"
        ]
        Resource = "${aws_cloudwatch_log_group.test.arn}:*"
      },
      {
        Effect = "Allow"
        Action = ["s3:*"]
        Resource = [
          aws_s3_bucket.bucket.arn,
          "${aws_s3_bucket.bucket.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_s3_bucket" "bucket" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_kinesis_stream" "stream" {
  name                      = var.rName
  encryption_type           = "NONE"
  enforce_consumer_deletion = false
  max_record_size_in_kib    = 1024
  retention_period          = 24
  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = "/aws/kinesis/${var.rName}"
}

data "aws_caller_identity" "current" {}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
