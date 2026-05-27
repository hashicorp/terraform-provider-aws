# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_route_table" "test" {
  count  = var.resource_count
  vpc_id = aws_vpc.test.id

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
