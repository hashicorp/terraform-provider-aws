# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = var.rName
}

resource "aws_globalaccelerator_custom_routing_listener" "test" {
  accelerator_arn = aws_globalaccelerator_custom_routing_accelerator.test.arn

  port_range {
    from_port = 443
    to_port   = 443
  }
}

resource "aws_globalaccelerator_custom_routing_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_custom_routing_listener.test.arn

  destination_configuration {
    from_port = 443
    to_port   = 8443
    protocols = ["TCP"]
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
