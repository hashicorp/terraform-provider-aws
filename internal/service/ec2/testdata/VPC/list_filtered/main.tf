# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_vpc" "expected" {
  count = 2

  cidr_block = "10.1.0.0/16"

  tags = {
    expected = "true"
  }
}

resource "aws_vpc" "not_expected" {
  count = 2

  cidr_block = "10.1.0.0/16"

  tags = {
    expected = "false"
  }
}
