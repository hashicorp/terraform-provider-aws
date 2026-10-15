# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_launch_template" "test" {
  count = var.resource_count
  name  = "${var.rName}-${count.index}"
}

variable "rName" {
  type     = string
  nullable = false
}

variable "resource_count" {
  type     = number
  nullable = false
}
