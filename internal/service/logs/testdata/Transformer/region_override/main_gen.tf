# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_transformer" "test" {
  region = var.region

  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }
}

resource "aws_cloudwatch_log_group" "test" {
  region = var.region

  name = var.rName
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
