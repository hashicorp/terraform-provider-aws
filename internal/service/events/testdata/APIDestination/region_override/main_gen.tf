# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_event_api_destination" "test" {
  region = var.region

  name                = var.rName
  invocation_endpoint = "https://example.com/"
  http_method         = "GET"
  connection_arn      = aws_cloudwatch_event_connection.test.arn
}

resource "aws_cloudwatch_event_connection" "test" {
  region = var.region

  name               = "${var.rName}-conn"
  authorization_type = "API_KEY"
  auth_parameters {
    api_key {
      key   = "testKey"
      value = "testValue"
    }
  }
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
