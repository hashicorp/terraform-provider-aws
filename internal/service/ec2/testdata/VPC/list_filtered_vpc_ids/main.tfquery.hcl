# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_vpc" "test" {
  provider = aws

  config {
    vpc_ids = local.vpc_ids
    filter {
      name   = "tag:expected"
      values = [var.rName]
    }
  }
}

locals {
  vpc_ids = concat(
    aws_vpc.expected[*].id,
    aws_vpc.not_expected[*].id,
  )
}
