# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iot_authorizer" "test" {
  name                    = var.rName
  authorizer_function_arn = aws_lambda_function.test.arn
  token_key_name          = "Token-Header-1"
  token_signing_public_keys = {
    Key1 = file("test-fixtures/iot-authorizer-signing-key.pem")
  }

  tags = var.resource_tags
}

resource "aws_iam_role" "test" {
  name = var.rName
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Principal = { Service = "lambda.amazonaws.com" }
      Effect    = "Allow"
    }]
  })
}

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = var.rName
  role             = aws_iam_role.test.arn
  handler          = "exports.example"
  runtime          = "nodejs24.x"
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
