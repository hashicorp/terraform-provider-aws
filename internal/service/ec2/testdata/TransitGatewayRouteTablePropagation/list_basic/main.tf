# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = var.rName
  }
}

resource "aws_vpc" "test" {
  count = var.resource_count

  cidr_block = cidrsubnet("10.0.0.0/8", 8, count.index)
}

resource "aws_subnet" "test" {
  count = var.resource_count

  vpc_id            = aws_vpc.test[count.index].id
  cidr_block        = cidrsubnet(aws_vpc.test[count.index].cidr_block, 8, 0)
  availability_zone = data.aws_availability_zones.available.names[count.index]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  count = var.resource_count

  subnet_ids         = [aws_subnet.test[count.index].id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test[count.index].id
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

resource "aws_ec2_transit_gateway_route_table_propagation" "test" {
  count = var.resource_count

  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test[count.index].id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of Transit Gateway route table propagations to create"
  type        = number
  nullable    = false
}
