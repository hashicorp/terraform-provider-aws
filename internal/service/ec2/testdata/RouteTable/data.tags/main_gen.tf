# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "aws_route_table" "test" {
  route_table_id = aws_route_table.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = var.resource_tags
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
