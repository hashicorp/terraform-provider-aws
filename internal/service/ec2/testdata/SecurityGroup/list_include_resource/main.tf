# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_security_group" "test" {
  count = var.resource_count

  name   = "${var.rName}-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "${var.rName}-${count.index}"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
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
