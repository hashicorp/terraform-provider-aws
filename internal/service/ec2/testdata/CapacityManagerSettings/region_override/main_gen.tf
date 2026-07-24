# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_capacity_manager_settings" "test" {
  region = var.region

  enabled = true
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
