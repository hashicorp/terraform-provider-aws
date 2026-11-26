# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_gateway_rate_limit" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  rate_limit_id      = "wildcards"
  dimension_keys     = ["targetName", "toolName"]

  # Fully specific: both positions pinned.
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

  # Trailing wildcard.
  entries {
    dimensions = {
      targetName = "search-target"
      toolName   = "*"
    }

    requests {
      rate   = 50
      period = "second"
    }
  }

  # rate = 0 blocks all matching traffic. Meaningful, not unset.
  entries {
    dimensions = {
      targetName = "deprecated-target"
      toolName   = "*"
    }

    requests {
      rate   = 0
      period = "second"
    }
  }

  # Catch-all, plus a fractional rate and a connections block.
  entries {
    dimensions = {
      targetName = "*"
      toolName   = "*"
    }

    requests {
      rate   = 0.5
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
