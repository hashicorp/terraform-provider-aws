# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id       = aws_api_gateway_rest_api.test.id
  resource_id       = aws_api_gateway_resource.test.id
  http_method       = aws_api_gateway_method.test.http_method
  status_code       = aws_api_gateway_method_response.error.status_code
  selection_pattern = ".*"

  response_templates = {
    "application/json" = ""
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  response_parameters = {
    "method.response.header.Content-Type" = "integration.response.body.type"
  }
}

resource "aws_api_gateway_rest_api" "test" {
  name = var.rName
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"

  response_models = {
    "application/json" = "Error"
  }

  response_parameters = {
    "method.response.header.Content-Type" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  request_templates = {
    "application/json" = ""
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  type = "MOCK"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
