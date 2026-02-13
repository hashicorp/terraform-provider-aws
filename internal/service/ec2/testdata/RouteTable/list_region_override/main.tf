# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.0.0.0/16"
}

resource "aws_route_table" "test" {
  count = 2

  region = var.region
  vpc_id = aws_vpc.test.id
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
