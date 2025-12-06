# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_route_table" "test" {
  region = var.region

  vpc_id = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.1.0.0/16"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
