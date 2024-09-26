# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
