# Copyright (c) HashiCorp, Inc.
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
