# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_eventbridgev2_event_source" "test" {
  name          = var.rName
  event_bus_arn = aws_eventbridgev2_event_bus.test.arn

  configuration {
    aws_service_events_configuration {
      aws_service = "aws.s3"
    }
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_eventbridgev2_event_bus" "test" {
  name = var.rName
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
