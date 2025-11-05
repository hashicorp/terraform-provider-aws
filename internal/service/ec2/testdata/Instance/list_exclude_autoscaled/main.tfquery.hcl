# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_instance" "excluded" {
  provider = aws

  config {
    filter {
      name   = "tag:test-filter"
      values = [var.rName]
    }
  }
}

list "aws_instance" "included" {
  provider = aws

  config {
    filter {
      name   = "tag:test-filter"
      values = [var.rName]
    }
    include_auto_scaled = true
  }
}
