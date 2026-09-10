# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = var.rName
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    http {
      agentcore_runtime {
        arn       = aws_bedrockagentcore_agent_runtime.test.agent_runtime_arn
        qualifier = var.qualifier
      }
    }
  }
}

resource "aws_bedrockagentcore_gateway" "test" {
  name            = var.rName
  role_arn        = aws_iam_role.test.arn
  authorizer_type = "AWS_IAM"
}

resource "aws_bedrockagentcore_agent_runtime" "test" {
  agent_runtime_name = replace(var.rName, "-", "_")
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = var.container_uri
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}

data "aws_iam_policy_document" "test_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer"
    ]
    effect    = "Allow"
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

variable "container_uri" {
  type     = string
  nullable = false
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "qualifier" {
  type     = string
  nullable = true
  default  = null
}