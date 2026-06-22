# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_event_target" "test" {
  count = var.resource_count

  rule      = aws_cloudwatch_event_rule.test.name
  target_id = "${var.rName}-${count.index}"
  arn       = aws_sns_topic.test[count.index].arn
}

resource "aws_cloudwatch_event_rule" "test" {
  name                = var.rName
  schedule_expression = "rate(1 hour)"
}

resource "aws_sns_topic" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
}

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
