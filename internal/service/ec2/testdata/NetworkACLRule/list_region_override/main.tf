# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_network_acl_rule" "test" {
  count  = var.resource_count
  region = var.region

  network_acl_id = aws_network_acl.test.id
  rule_number    = 200 + count.index
  egress         = false
  protocol       = "6"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_network_acl" "test" {
  region = var.region

  vpc_id = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
