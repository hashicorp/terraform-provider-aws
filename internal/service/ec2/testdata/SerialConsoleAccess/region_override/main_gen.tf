# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_serial_console_access" "test" {
  region = var.region

  enabled = true
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
