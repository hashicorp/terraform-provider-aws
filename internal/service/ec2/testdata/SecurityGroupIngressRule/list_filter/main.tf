# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_security_group" "expected" {
  name   = "${var.rName}-expected"
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "not_expected" {
  name   = "${var.rName}-not-expected"
  vpc_id = aws_vpc.test.id
}

resource "aws_vpc_security_group_ingress_rule" "expected" {
  count = 2

  security_group_id = aws_security_group.expected.id
  cidr_ipv4         = "10.0.${count.index}.0/24"
  from_port         = 80 + count.index
  to_port           = 80 + count.index
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_ingress_rule" "not_expected" {
  count = 2

  security_group_id = aws_security_group.not_expected.id
  cidr_ipv4         = "10.1.${count.index}.0/24"
  from_port         = 443 + count.index
  to_port           = 443 + count.index
  ip_protocol       = "tcp"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
