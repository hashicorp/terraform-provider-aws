# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_subnet" "test" {
  provider = aws

  config {
    subnet_ids = local.subnet_ids
    filter {
      name   = "tag:expected"
      values = [var.rName]
    }
  }
}

locals {
  subnet_ids = concat(
    aws_subnet.expected[*].id,
    aws_subnet.not_expected[*].id
  )
}
