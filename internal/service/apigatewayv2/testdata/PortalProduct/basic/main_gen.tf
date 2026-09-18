# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_apigatewayv2_portal_product" "test" {
  display_name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
