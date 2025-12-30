# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_subnet" "test" {
  count = length(aws_vpc.test) * 2

  cidr_block = cidrsubnet(aws_vpc.test[floor(count.index / 2)].cidr_block, 8, count.index)
  vpc_id     = aws_vpc.test[floor(count.index / 2)].id
}

resource "aws_vpc" "test" {
  count = 2

  cidr_block = "10.1.0.0/16"
}
