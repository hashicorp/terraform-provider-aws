# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_alarm_mute_rule" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"

  rule {
    schedule {
      duration   = "PT4H"
      expression = "cron(0 2 * * *)"
    }
  }
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
