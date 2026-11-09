# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_stream" "test" {
  region = var.region

  name           = "${var.rName}-s"
  log_group_name = aws_cloudwatch_log_group.test.id
}

resource "aws_cloudwatch_log_group" "test" {
  region = var.region

  name = "${var.rName}-g"
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
