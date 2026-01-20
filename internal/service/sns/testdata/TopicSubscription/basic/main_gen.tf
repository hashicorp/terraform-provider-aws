# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_sns_topic" "test" {
  name = var.rName
}

resource "aws_sqs_queue" "test" {
  name = var.rName

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
