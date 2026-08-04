# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_scheduler_schedule" "test" {
  count  = var.resource_count
  region = var.region

  name = "${var.rName}-${count.index}"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}

resource "aws_sqs_queue" "test" {
  region = var.region
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "scheduler.${data.aws_partition.main.dns_suffix}"
      }
      Condition = {
        StringEquals = {
          "aws:SourceAccount" : data.aws_caller_identity.main.account_id
        }
      }
    }
  })
}

data "aws_caller_identity" "main" {}

data "aws_partition" "main" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
