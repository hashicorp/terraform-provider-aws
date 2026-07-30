# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_event_archive" "test" {
  name             = var.rName
  event_source_arn = aws_cloudwatch_event_bus.test.arn
  retention_days   = var.retention_days
  description      = var.description
  event_pattern    = <<PATTERN
{
  "source": ["company.team.service"]
}
PATTERN
}

resource "aws_cloudwatch_event_bus" "test" {
  name = "${var.rName}-bus"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "description" {
  type     = string
  nullable = false
}

variable "retention_days" {
  type     = number
  nullable = false
}