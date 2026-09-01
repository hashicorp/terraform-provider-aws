# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_globalaccelerator_accelerator" "test" {
  name            = var.rName
  ip_address_type = "IPV4"
  enabled         = false
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
