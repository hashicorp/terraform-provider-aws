# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_vpc" "test" {
  count = 3

  cidr_block = "10.1.0.0/16"
}
