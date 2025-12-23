# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
