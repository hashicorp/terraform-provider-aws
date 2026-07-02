# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0


resource "aws_vpc" "expected" {
  count = 2

  cidr_block = "10.1.0.0/16"

  tags = {
    expected = var.rName
  }
}

resource "aws_vpc" "not_expected" {
  count = 2

  cidr_block = "10.1.0.0/16"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
