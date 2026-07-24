# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_delivery_destination" "test" {
  name = var.rName

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.test.arn
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
