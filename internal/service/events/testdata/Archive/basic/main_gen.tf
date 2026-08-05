# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_event_archive" "test" {
  name             = var.rName
  event_source_arn = aws_cloudwatch_event_bus.test.arn
}

resource "aws_cloudwatch_event_bus" "test" {
  name = "${var.rName}-bus"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
