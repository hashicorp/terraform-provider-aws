# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_apigatewayv2_integration" "test" {
  count = var.resource_count

  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "MOCK"

  description            = "test-integration"
  passthrough_behavior   = "WHEN_NO_MATCH"
  payload_format_version = "1.0"
  timeout_milliseconds   = 29000
}

resource "aws_apigatewayv2_api" "test" {
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