# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_api_gateway_method" "test" {
  count = length(var.http_methods)

  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = var.http_methods[count.index]
  authorization = "NONE"
}

resource "aws_api_gateway_rest_api" "test" {
  name = var.rName
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "http_methods" {
  description = "HTTP methods to create"
  type        = list(string)
  nullable    = false
}
