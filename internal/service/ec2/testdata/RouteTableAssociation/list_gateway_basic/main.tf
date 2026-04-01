# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route_table_association" "internet" {
  route_table_id = aws_route_table.test.id
  gateway_id     = aws_internet_gateway.test.id
}

resource "aws_route_table_association" "vpn" {
  route_table_id = aws_route_table.test.id
  gateway_id     = aws_vpn_gateway.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block           = aws_subnet.test.cidr_block
    network_interface_id = aws_network_interface.test.id
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

# testAccRouteTableAssociationConfigBaseVPC

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
