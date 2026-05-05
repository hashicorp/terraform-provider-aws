# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route_table_association" "test" {
  count = var.resource_count

  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test[count.index].id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.test.id
  }
}

# testAccRouteTableAssociationConfigBaseVPC

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  count = var.resource_count

  vpc_id     = aws_vpc.test.id
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
