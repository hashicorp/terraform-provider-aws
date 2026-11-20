# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_xray_trace_segment_destination" "test" {
  destination = "CloudWatchLogs"

  cloudwatch_logs_configuration {
    spans_log_group_kms_key_id = aws_kms_key.test.arn
  }

  depends_on = [aws_cloudwatch_log_resource_policy.test, aws_kms_key_policy.test]
}

data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_kms_key" "test" {
  description             = var.rName
  deletion_window_in_days = 7
}

resource "aws_kms_key_policy" "test" {
  key_id = aws_kms_key.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "EnableRoot"
        Effect    = "Allow"
        Principal = { AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root" }
        Action    = "kms:*"
        Resource  = "*"
      },
      {
        Sid       = "AllowLogs"
        Effect    = "Allow"
        Principal = { Service = "logs.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}" }
        Action    = ["kms:Encrypt*", "kms:Decrypt*", "kms:ReEncrypt*", "kms:GenerateDataKey*", "kms:Describe*"]
        Resource  = "*"
      }
    ]
  })
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:PutLogEvents",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:logs:*:*:log-group:aws/spans:*",
      "arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/application-signals/data:*"
    ]

    principals {
      identifiers = ["xray.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = var.rName
  policy_document = data.aws_iam_policy_document.test.json
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
