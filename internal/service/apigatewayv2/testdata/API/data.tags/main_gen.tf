# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_apigatewayv2_api" "test" {
  api_id = aws_apigatewayv2_api.test.id
}

resource "aws_apigatewayv2_api" "test" {
  name                       = var.rName
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
