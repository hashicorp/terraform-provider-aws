# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpc_security_group_vpc_association" "test" {
  region = var.region

  security_group_id = aws_security_group.test.id
  vpc_id            = aws_vpc.target.id
}

resource "aws_vpc" "source" {
  region = var.region

  cidr_block = "10.6.0.0/16"
}

resource "aws_security_group" "test" {
  region = var.region

  name   = var.rName
  vpc_id = aws_vpc.source.id
}

resource "aws_vpc" "target" {
  region = var.region

  cidr_block = "10.7.0.0/16"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
