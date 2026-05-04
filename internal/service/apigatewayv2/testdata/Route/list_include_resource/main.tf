# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_apigatewayv2_route" "test" {
  count = var.resource_count

  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$connect"

  api_key_required                    = true
  operation_name                      = "GET"
  route_response_selection_expression = "$default"
  model_selection_expression          = "action"
  authorization_type                  = "NONE"

  request_parameter {
    request_parameter_key = "route.request.header.authorization"
    required              = true
  }
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
