# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_event_permission" "test" {
  principal      = "111111111111"
  statement_id   = var.rName
  event_bus_name = aws_cloudwatch_event_bus.test.name
}

resource "aws_cloudwatch_event_bus" "test" {
  name = "${var.rName}-bus"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
