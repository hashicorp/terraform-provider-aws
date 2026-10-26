# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_stream" "test" {
  count = var.resource_count

  name           = "${var.rName}-${count.index}"
  log_group_name = aws_cloudwatch_log_group.test.id
}

resource "aws_cloudwatch_log_group" "test" {
  name = "${var.rName}-g"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
