# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_route53_zone" "public" {
  name = "public.${var.zoneName}"
}

resource "aws_route53_zone" "private" {
  name = "private.${var.zoneName}"

  vpc {
    vpc_id = aws_vpc.test.id
  }
}

variable "zoneName" {
  description = "Root zone name for hosted zones"
  type        = string
  nullable    = false
}