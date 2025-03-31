# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_vpclattice_resource_gateway" "test" {
  name       = var.rName
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 0)

  tags = {
    Name = var.rName
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
