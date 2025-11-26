# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

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
