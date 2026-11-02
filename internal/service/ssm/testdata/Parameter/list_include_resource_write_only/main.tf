# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_parameter" "test" {
  count = var.resource_count

  name             = "${var.rName}-${count.index}"
  type             = "SecureString"
  value_wo         = "${var.rName}-${count.index}"
  value_wo_version = 1

  tags = {
    Name = "${var.rName}-${count.index}"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
