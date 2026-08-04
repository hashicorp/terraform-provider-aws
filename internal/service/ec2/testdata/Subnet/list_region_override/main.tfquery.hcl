# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_subnet" "test" {
  provider = aws

  config {
    region = var.region
  }
}
