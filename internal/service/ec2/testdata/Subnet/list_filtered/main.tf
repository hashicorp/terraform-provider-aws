# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

# provider "aws" {}

resource "aws_subnet" "expected" {
  count = 2

  cidr_block = cidrsubnet(aws_vpc.expected.cidr_block, 8, count.index)
  vpc_id     = aws_vpc.expected.id
}

resource "aws_vpc" "expected" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "not_expected" {
  count = 2

  cidr_block = cidrsubnet(aws_vpc.not_expected.cidr_block, 8, count.index)
  vpc_id     = aws_vpc.not_expected.id
}

resource "aws_vpc" "not_expected" {
  cidr_block = "10.1.0.0/16"
}
