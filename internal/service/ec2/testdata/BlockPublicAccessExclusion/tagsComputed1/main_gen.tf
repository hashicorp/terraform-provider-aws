# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_vpc_block_public_access_exclusion" "test" {
  internet_gateway_exclusion_mode = "allow-bidirectional"
  vpc_id                          = aws_vpc.test.id

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "null_resource" "test" {}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
