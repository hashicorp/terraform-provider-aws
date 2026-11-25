# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_public_ipv4_pool" "test" {
  count = var.resource_count
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
