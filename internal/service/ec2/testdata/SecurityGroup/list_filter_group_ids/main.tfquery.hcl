# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_security_group" "test" {
  provider = aws

  config {
    group_ids = [
      aws_security_group.expected[0].id,
      aws_security_group.expected[1].id,
    ]
  }
}
