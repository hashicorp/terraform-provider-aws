# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_security_group" "expected" {
  count = 2

  vpc_id = aws_vpc.test.id
  name   = "${var.rName}-expected-${count.index}"

  tags = {
    Name = "${var.rName}-expected-${count.index}"
  }
}

resource "aws_security_group" "not_expected" {
  count = 2

  vpc_id = aws_vpc.test.id
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
