# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_eventbridgev2_event_source" "test" {
  count = var.resource_count

  name          = "${var.rName}-${count.index}"
  event_bus_arn = aws_eventbridgev2_event_bus.test.arn

  configuration {
    aws_service_events_configuration {
      aws_service = "aws.s3"
    }
  }
}

resource "aws_eventbridgev2_event_bus" "test" {
  name = var.rName
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
