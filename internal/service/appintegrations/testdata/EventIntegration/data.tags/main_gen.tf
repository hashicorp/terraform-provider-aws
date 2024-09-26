# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "aws_appintegrations_event_integration" "test" {
  name = aws_appintegrations_event_integration.test.name
}

resource "aws_appintegrations_event_integration" "test" {
  name            = var.rName
  eventbridge_bus = "default"

  event_filter {
    source = "aws.partner/examplepartner.com"
  }

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
