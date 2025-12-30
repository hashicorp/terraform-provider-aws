# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_globalaccelerator_accelerator" "example" {
  name            = var.rName
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "test" {
  accelerator_arn = aws_globalaccelerator_accelerator.example.arn
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 81
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
