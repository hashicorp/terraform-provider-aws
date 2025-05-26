# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_serial_console_access" "test" {
  enabled = true
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
