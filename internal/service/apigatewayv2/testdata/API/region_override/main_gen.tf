# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_apigatewayv2_api" "test" {
  region = var.region

  name                       = var.rName
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
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
