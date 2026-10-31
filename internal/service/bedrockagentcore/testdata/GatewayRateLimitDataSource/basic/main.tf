# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_bedrockagentcore_gateway_rate_limit" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway_rate_limit.test.gateway_identifier
  rate_limit_id      = aws_bedrockagentcore_gateway_rate_limit.test.rate_limit_id
}

resource "aws_bedrockagentcore_gateway_rate_limit" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  rate_limit_id      = "ds-per-target"
  description        = "data source acceptance test"
  dimension_keys     = ["targetName", "toolName"]

  entries {
    dimensions = {
      targetName = "search-target"
      toolName   = "readData"
    }

    requests {
      rate   = 250
      period = "second"
    }
  }

  entries {
    dimensions = {
      targetName = "*"
      toolName   = "*"
    }

    tokens {
      rate   = 50000
      period = "minute"
    }

    connections {
      rate   = 5
      period = "second"
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
