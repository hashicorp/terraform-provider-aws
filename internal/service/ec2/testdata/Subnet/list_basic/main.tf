# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_subnet" "test" {
  count = length(aws_vpc.test)

  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test[count.index].id
}

resource "aws_vpc" "test" {
  count = 3

  cidr_block = "10.1.0.0/16"
}

