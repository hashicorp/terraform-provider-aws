# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_subnet" "test" {
  provider = aws

  config {
    subnet_ids = local.subnet_ids
  }
}

locals {
  subnet_ids = aws_subnet.test[*].id
}
