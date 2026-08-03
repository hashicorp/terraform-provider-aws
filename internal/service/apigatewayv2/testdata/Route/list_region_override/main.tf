# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_apigatewayv2_route" "test" {
  count  = var.resource_count
  region = var.region

  api_id    = aws_apigatewayv2_api.test.id
  route_key = "test-${count.index}"
}

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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to create resources in"
  type        = string
  nullable    = false
}
