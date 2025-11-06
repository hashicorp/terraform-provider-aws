# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_vpc" "test" {
  provider = aws

  config {
    region = var.region
  }
}
