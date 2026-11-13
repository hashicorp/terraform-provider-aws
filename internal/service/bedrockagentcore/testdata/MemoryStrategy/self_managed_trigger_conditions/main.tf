# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_memory_strategy" "test" {
  name      = "${var.rName}_s"
  memory_id = aws_bedrockagentcore_memory.test.id
  type      = "CUSTOM"

  configuration {
    type = "SELF_MANAGED"

    self_managed_configuration {
      invocation_configuration {
        payload_delivery_bucket_name = aws_s3_bucket.test.bucket
        topic_arn                    = aws_sns_topic.test.arn
      }

      trigger_conditions {
        dynamic "message_based_trigger" {
          for_each = var.message_count != null ? [var.message_count] : []

          content {
            message_count = message_based_trigger.value
          }
        }

        dynamic "time_based_trigger" {
          for_each = var.idle_session_timeout != null ? [var.idle_session_timeout] : []

          content {
            idle_session_timeout = time_based_trigger.value
          }
        }

        dynamic "token_based_trigger" {
          for_each = var.token_count != null ? [var.token_count] : []

          content {
            token_count = token_based_trigger.value
          }
        }
      }
    }
  }
}

resource "aws_sns_topic" "test" {
  name = "${var.rName}-t"
}

resource "aws_s3_bucket" "test" {
  bucket        = "${replace(var.rName, "_", "-")}-b"
  force_destroy = false
}

resource "aws_bedrockagentcore_memory" "test" {
  name                      = "${var.rName}_m"
  event_expiry_duration     = 7
  memory_execution_role_arn = aws_iam_role.test.arn
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = "${var.rName}-role"
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/BedrockAgentCoreFullAccess"
}

resource "aws_iam_policy" "sns_topic_full_access" {
  name        = "${var.rName}-sns-topic"
  description = "Full SNS permissions for one topic"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "sns:*"
      Resource = aws_sns_topic.test.arn
    }]
  })
}

resource "aws_iam_role_policy_attachment" "sns_topic_full_access" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.sns_topic_full_access.arn
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

variable "message_count" {
  type     = number
  nullable = true
  default  = null
}

variable "idle_session_timeout" {
  type     = number
  nullable = true
  default  = null
}

variable "token_count" {
  type     = number
  nullable = true
  default  = null
}