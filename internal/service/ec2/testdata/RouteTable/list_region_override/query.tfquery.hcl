# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_route_table" "test" {
  provider = aws

  config {
    region = var.region

    filter {
      name   = "vpc-id"
      values = [aws_vpc.test.id]
    }
  }
}
