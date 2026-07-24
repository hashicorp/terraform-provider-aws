# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_resource_policy" "test" {
  count = var.resource_count

  resource_arn = aws_bedrockagentcore_agent_runtime.test[count.index].agent_runtime_arn
  policy       = data.aws_iam_policy_document.resource_policy[count.index].json
}

resource "aws_bedrockagentcore_agent_runtime" "test" {
  count = var.resource_count

  agent_runtime_name = "${var.rName}_${count.index}"
  role_arn           = aws_iam_role.test.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = var.AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}

data "aws_iam_policy_document" "resource_policy" {
  count = var.resource_count

  statement {
    effect = "Allow"
    actions = [
      "bedrock-agentcore:InvokeAgentRuntime",
    ]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    resources = [
      aws_bedrockagentcore_agent_runtime.test[count.index].agent_runtime_arn
    ]
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

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI" {
  type     = string
  nullable = false
}
