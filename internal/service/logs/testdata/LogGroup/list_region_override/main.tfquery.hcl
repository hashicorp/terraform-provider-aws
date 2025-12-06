# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

list "aws_cloudwatch_log_group" "test" {
  provider = aws

  config {
    region = var.region
  }
}
