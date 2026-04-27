# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_internet_gateway" "expected" {
  count = 2

  tags = {
    Name  = "${var.rName}-expected-${count.index}"
    Scope = "expected"
  }
}

resource "aws_internet_gateway" "other" {
  tags = {
    Name  = "${var.rName}-other"
    Scope = "other"
  }
}

variable "rName" {
  description = "Name prefix for resources"
  type        = string
  nullable    = false
}