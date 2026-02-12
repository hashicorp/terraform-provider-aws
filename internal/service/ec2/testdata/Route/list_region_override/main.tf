# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_route" "test" {
  count = 2

  region                 = var.region
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "172.${16 + count.index}.0.0/16"
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_route_table" "test" {
  region = var.region
  vpc_id = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.0.0.0/16"
}

resource "aws_internet_gateway" "test" {
  region = var.region
  vpc_id = aws_vpc.test.id
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
