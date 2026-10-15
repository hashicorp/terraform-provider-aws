# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_cloudwatch_event_rule" "test" {
  provider = aws

  config {
    region = var.region
  }
}
