# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route" "test" {
  count = 2

  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "172.${16 + count.index}.0.0/16"
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
