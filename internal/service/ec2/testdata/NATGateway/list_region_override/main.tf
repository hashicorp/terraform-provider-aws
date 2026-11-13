# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_nat_gateway" "test" {
  count  = var.resource_count
  region = var.region

  connectivity_type = "private"
  subnet_id         = aws_subnet.test[count.index].id

  tags = {
    Name = "${var.rName}-${count.index}"
  }
}

resource "aws_vpc" "test" {
  region     = var.region
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count  = var.resource_count
  region = var.region

  vpc_id                  = aws_vpc.test.id
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  map_public_ip_on_launch = false

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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
