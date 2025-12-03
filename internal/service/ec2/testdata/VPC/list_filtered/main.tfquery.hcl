# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_vpc" "test" {
  provider = aws

  config {
    filter {
      name   = "tag:expected"
      values = [var.rName]
    }
  }
}
