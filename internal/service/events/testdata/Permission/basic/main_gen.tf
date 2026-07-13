# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_event_permission" "test" {
  principal    = "111111111111"
  statement_id = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
