# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_rest_api" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
