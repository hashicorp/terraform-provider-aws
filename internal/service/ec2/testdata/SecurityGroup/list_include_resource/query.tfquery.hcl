# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_security_group" "test" {
  provider = aws

  include_resource = true

  config {
    filter {
      name   = "vpc-id"
      values = [aws_vpc.test.id]
    }
  }
}
