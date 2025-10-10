# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_vpc" "test" {
  provider = aws

  config {
    vpc_ids = local.vpc_ids
  }
}

locals {
  vpc_ids = aws_vpc.test[*].id
}
