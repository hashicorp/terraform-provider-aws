# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_gateway_rate_limit" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
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

# Same dimension_keys tuple on the same gateway. The API rejects this with a
# ValidationException -- NOT the ConflictException the AWS docs claim.
resource "aws_bedrockagentcore_gateway_rate_limit" "duplicate" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  dimension_keys     = ["targetName"]

  entries {
    dimensions = {
      targetName = "*"
    }

    requests {
      rate   = 1
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
