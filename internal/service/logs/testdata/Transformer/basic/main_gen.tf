# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = var.rName
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
