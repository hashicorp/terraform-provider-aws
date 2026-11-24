# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatchomni_view" "test" {
  name       = "view.${var.rName}"
  definition = var.definition
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "definition" {
  type     = string
  nullable = false
}