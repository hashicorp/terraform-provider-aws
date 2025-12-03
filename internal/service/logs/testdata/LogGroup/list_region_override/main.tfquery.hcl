# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_cloudwatch_log_group" "test" {
  provider = aws

  config {
    region = var.region
  }
}
