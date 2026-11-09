# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_transit_gateway_route_table_propagation" "test" {
  count  = var.resource_count
  region = var.region

  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test[count.index].id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}

resource "aws_ec2_transit_gateway" "test" {
  region = var.region

  tags = {
    Name = var.rName
  }
}

resource "aws_vpc" "test" {
  count  = var.resource_count
  region = var.region

  cidr_block = cidrsubnet("10.0.0.0/8", 8, count.index)
}

resource "aws_subnet" "test" {
  count  = var.resource_count
  region = var.region

  vpc_id            = aws_vpc.test[count.index].id
  cidr_block        = cidrsubnet(aws_vpc.test[count.index].cidr_block, 8, 0)
  availability_zone = data.aws_availability_zones.available.names[count.index]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  count  = var.resource_count
  region = var.region

  subnet_ids         = [aws_subnet.test[count.index].id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test[count.index].id
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  region = var.region

  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

data "aws_availability_zones" "available" {
  region = var.region
  state  = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
