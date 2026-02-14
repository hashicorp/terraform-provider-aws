# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_subnet" "expected" {
  count = 2

  cidr_block = cidrsubnet(aws_vpc.test[count.index].cidr_block, 8, 0)
  vpc_id     = aws_vpc.test[count.index].id

  tags = {
    expected = var.rName
  }
}

resource "aws_subnet" "not_expected" {
  count = 2

  cidr_block = cidrsubnet(aws_vpc.test[count.index].cidr_block, 8, 1)
  vpc_id     = aws_vpc.test[count.index].id
}

resource "aws_vpc" "test" {
  count = 2

  cidr_block = "10.1.0.0/16"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
