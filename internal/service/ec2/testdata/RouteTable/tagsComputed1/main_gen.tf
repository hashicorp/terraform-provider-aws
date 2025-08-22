# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "null_resource" "test" {}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
