# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_event_api_destination" "test" {
  name                = var.rName
  invocation_endpoint = "https://example.com/"
  http_method         = "GET"
  connection_arn      = aws_cloudwatch_event_connection.test.arn
}

resource "aws_cloudwatch_event_connection" "test" {
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
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.53.0"
    }
  }
}

provider "aws" {}
