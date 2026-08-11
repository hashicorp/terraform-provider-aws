# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_eip" "test" {
  count = var.resource_count

  domain = "vpc"
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
