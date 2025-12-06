# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_event_target" "test" {
  rule      = aws_cloudwatch_event_rule.test.name
  target_id = var.rName
  arn       = aws_sns_topic.test.arn
}

resource "aws_cloudwatch_event_rule" "test" {
  name                = var.rName
  schedule_expression = "rate(1 hour)"
}

resource "aws_sns_topic" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
