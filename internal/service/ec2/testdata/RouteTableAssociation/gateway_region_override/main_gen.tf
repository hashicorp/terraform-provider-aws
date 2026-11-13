# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route_table_association" "test" {
  region = var.region

  route_table_id = aws_route_table.test.id
  gateway_id     = aws_vpn_gateway.test.id
}

resource "aws_route_table" "test" {
  region = var.region

  vpc_id = aws_vpc.test.id

  route {
    cidr_block           = aws_subnet.test.cidr_block
    network_interface_id = aws_network_interface.test.id
  }
}

resource "aws_network_interface" "test" {
  region = var.region

  subnet_id = aws_subnet.test.id
}

resource "aws_vpn_gateway" "test" {
  region = var.region

  vpc_id = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  region = var.region

  vpc_id     = aws_vpc.test.id
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
