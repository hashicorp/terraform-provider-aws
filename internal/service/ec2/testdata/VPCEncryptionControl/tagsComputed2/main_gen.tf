# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_vpc_encryption_control" "test" {
  vpc_id = aws_vpc.test.id
  mode   = "monitor"

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
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

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
