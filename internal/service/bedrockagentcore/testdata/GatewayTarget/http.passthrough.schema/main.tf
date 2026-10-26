# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = var.rName
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {
      service = "lambda"
    }
  }

  target_configuration {
    http {
      passthrough {
        endpoint      = "https://example.com/api"
        protocol_type = "MCP"

        schema {
          source {
            inline_payload {
              payload = jsonencode({
                openapi = "3.0.0"
                info = {
                  title   = "Test API"
                  version = "1.0.0"
                }
                paths = {}
              })
            }
          }
        }
      }
    }
  }
}

resource "aws_bedrockagentcore_gateway" "test" {
  name            = var.rName
  role_arn        = aws_iam_role.test.arn
  authorizer_type = "AWS_IAM"
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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
