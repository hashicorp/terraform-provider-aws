# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_instance" "test" {
  provider = aws

  config {
    filter {
      name   = "instance-state-name"
      values = ["stopped"]
    }
  }
}
