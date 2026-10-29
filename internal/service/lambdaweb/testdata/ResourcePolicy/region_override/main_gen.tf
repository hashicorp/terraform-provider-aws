# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_lambdaweb_resource_policy" "test" {
  region = var.region

  resource_arn = aws_lambdaweb_function.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "AllowAccount"
      Effect = "Allow"
      Principal = {
        AWS = data.aws_caller_identity.current.account_id
      }
      Action   = "lambda:InvokeWebFunction"
      Resource = aws_lambdaweb_function.test.arn
    }]
  })
}

resource "aws_lambdaweb_function" "test" {
  region = var.region

  function_name = var.rName
}

data "aws_caller_identity" "current" {}

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
