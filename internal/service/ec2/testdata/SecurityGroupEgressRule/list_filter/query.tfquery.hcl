# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_vpc_security_group_egress_rule" "test" {
  provider = aws

  config {
    filter {
      name   = "group-id"
      values = [aws_security_group.expected.id]
    }
  }
}
