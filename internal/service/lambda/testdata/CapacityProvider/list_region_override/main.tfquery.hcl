# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_lambda_capacity_provider" "test" {
  provider = aws

  config {
    region = var.region
  }
}