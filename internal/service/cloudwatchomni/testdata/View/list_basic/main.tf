# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatchomni_view" "test" {
  count = var.resource_count

  name       = "view.${var.rName}-${count.index}"
  definition = "SELECT `@timestamp`, `@message` FROM \"logs.default\""
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
