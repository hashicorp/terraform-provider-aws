# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_neptunegraph_graph" "test" {
  region              = var.region
  graph_name          = var.rName
  provisioned_memory  = 16
  deletion_protection = false
}

resource "aws_neptunegraph_private_graph_endpoint" "test" {
  region           = var.region
  graph_identifier = aws_neptunegraph_graph.test.id
  vpc_id           = aws_vpc.test.id
  subnet_ids       = aws_subnet.test[*].id
}

resource "aws_vpc" "test" {
  region     = var.region
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_subnet" "test" {
  count  = 2
  region = var.region

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = var.rName
  }
}

data "aws_availability_zones" "available" {
  region = var.region
  state  = "available"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
