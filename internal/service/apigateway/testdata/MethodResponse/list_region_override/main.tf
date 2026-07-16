# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_api_gateway_method_response" "test" {
  count  = length(var.status_codes)
  region = var.region

  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = var.status_codes[count.index]
}

resource "aws_api_gateway_rest_api" "test" {
  region = var.region

  name = var.rName
}

resource "aws_api_gateway_resource" "test" {
  region = var.region

  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  region = var.region

  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
