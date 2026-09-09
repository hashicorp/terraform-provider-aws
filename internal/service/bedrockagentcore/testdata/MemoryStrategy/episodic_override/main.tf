# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = "${var.rName}_s"
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = "CUSTOM"
  namespace_templates = [var.namespace_template]

  configuration {
    type = "EPISODIC_OVERRIDE"

    consolidation {
      append_to_prompt = "<task>Consolidate</task>"
      model_id         = "us.amazon.nova-2-lite-v1:0"
    }

    reflection {
      append_to_prompt    = "<task>Reflect</task>"
      model_id            = "amazon.nova-lite-v1:0"
      namespace_templates = [var.reflection_namespace_template]
    }
  }
}

resource "aws_bedrockagentcore_memory" "test" {
  name                      = "${var.rName}_m"
  event_expiry_duration     = 7
  memory_execution_role_arn = aws_iam_role.test.arn
}

data "aws_partition" "current" {}

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

resource "aws_iam_role" "test" {
  name               = "${var.rName}-role"
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/BedrockAgentCoreFullAccess"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "namespace_template" {
  type     = string
  nullable = false
}

variable "reflection_namespace_template" {
  type     = string
  nullable = false
}