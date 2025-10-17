# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_subnet" "test" {
  provider = aws

  config {
    filter {
      name   = "vpc-id"
      values = [aws_vpc.expected.id]
    }
  }
}
