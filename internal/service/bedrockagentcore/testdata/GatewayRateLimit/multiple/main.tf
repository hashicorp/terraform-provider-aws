# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_gateway_rate_limit" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  rate_limit_id      = "per-target"
  dimension_keys     = ["targetName"]

  entries {
    dimensions = {
      targetName = "*"
    }

    requests {
      rate   = 100
      period = "second"
    }
  }
}

# Same gateway, different dimension_keys. Serialised by the shared per-gateway
# mutex; tokens here is minute-only per the API.
resource "aws_bedrockagentcore_gateway_rate_limit" "second" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  dimension_keys     = ["$.context.jwt.sub"]

  entries {
    dimensions = {
      "$.context.jwt.sub" = "*"
    }

    tokens {
      rate   = 50000
      period = "minute"
    }
  }
}

data "aws_iam_policy_document" "gateway_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "gateway" {
  name               = "${var.rName}-gateway"
  assume_role_policy = data.aws_iam_policy_document.gateway_assume.json
}

resource "aws_bedrockagentcore_gateway" "test" {
  name     = "${var.rName}-gateway"
  role_arn = aws_iam_role.gateway.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test"]
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
