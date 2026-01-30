# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_vpc" "expected" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "${var.rName}-expected"
  }
}

resource "aws_vpc" "not_expected" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "${var.rName}-not-expected"
  }
}

resource "aws_security_group" "expected" {
  count = 2

  vpc_id = aws_vpc.expected.id
  name   = "${var.rName}-expected-${count.index}"

  tags = {
    Name = "${var.rName}-expected-${count.index}"
  }
}

resource "aws_security_group" "not_expected" {
  count = 2

  vpc_id = aws_vpc.not_expected.id
  name   = "${var.rName}-not-expected-${count.index}"

  tags = {
    Name = "${var.rName}-not-expected-${count.index}"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
