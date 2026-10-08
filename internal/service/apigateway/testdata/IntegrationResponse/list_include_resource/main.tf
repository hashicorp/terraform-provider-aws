# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_api_gateway_integration_response" "test" {
  count = length(var.status_codes)

  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = var.status_codes[count.index]

  depends_on = [aws_api_gateway_integration.test]
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
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type = "MOCK"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "status_codes" {
  description = "Status codes to create"
  type        = list(string)
  nullable    = false
}
