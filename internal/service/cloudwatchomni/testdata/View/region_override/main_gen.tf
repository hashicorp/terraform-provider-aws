# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatchomni_view" "test" {
  region = var.region

  name       = "view.${var.rName}"
  definition = "SELECT `@timestamp`, `@message` FROM \"logs.default\""
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
