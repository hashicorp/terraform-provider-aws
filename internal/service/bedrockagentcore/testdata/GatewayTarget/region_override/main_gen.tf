# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_gateway_target" "test" {
  region = var.region

  name               = var.rName
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    mcp {
      lambda {
        lambda_arn = aws_lambda_function.test.arn

        tool_schema {
          inline_payload {
            name        = "test_tool"
            description = "A test tool"

            input_schema {
              type = "object"

              property {
                name        = "input"
                description = "some input"
                type        = "string"
                required    = true
              }
            }
          }
        }
      }
    }
  }
}

resource "aws_bedrockagentcore_gateway" "test" {
  region = var.region

  name     = var.rName
  role_arn = aws_iam_role.test.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test"]
    }
  }

  protocol_configuration {
    mcp {
      instructions       = "Do something"
      supported_versions = ["2025-11-25"]
    }
  }

  protocol_type = "MCP"
}

data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "lambda:*",
    "Resource": "*"
  }
}
EOF
}

data "aws_iam_policy_document" "lambda_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name = "${var.rName}-lambda"

  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

resource "aws_lambda_function" "test" {
  region = var.region

  filename      = "test-fixtures/lambdatest.zip"
  function_name = var.rName
  role          = aws_iam_role.lambda.arn
  handler       = "lambdatest.handler"
  runtime       = "nodejs24.x"
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
