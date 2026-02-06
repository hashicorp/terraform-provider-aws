# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_vpc" "test" {
  provider = aws

  config {
    filter {
      name   = "is-default"
      values = ["false"]
    }
  }
}
