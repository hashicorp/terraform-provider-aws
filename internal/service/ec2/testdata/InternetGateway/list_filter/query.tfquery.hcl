# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_internet_gateway" "test" {
  provider = aws

  config {
    filter {
      name   = "internet-gateway-id"
      values = [aws_internet_gateway.expected[0].id, aws_internet_gateway.expected[1].id]
    }
  }
}