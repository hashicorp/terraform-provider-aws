# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = var.rName
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    mcp {
      connector {
        source {
          connector_id = "bedrock-knowledge-bases"
        }

        configuration {
          name = "Retrieve"

          parameter_values = jsonencode({
            knowledgeBaseId = aws_bedrockagent_knowledge_base.test.id
          })
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_bedrockagentcore_gateway" "test" {
  name            = var.rName
  role_arn        = aws_iam_role.test.arn
  authorizer_type = "AWS_IAM"
  protocol_type   = "MCP"
}

data "aws_caller_identity" "current" {}

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
    "Statement": [
      {
        "Sid": "ValidateKnowledgeBase",
        "Effect": "Allow",
        "Action": "bedrock:GetKnowledgeBase",
        "Resource": "${aws_bedrockagent_knowledge_base.test.arn}"
      },
      {
        "Sid": "RetrieveFromKnowledgeBase",
        "Effect": "Allow",
        "Action": "bedrock:Retrieve",
        "Resource": "${aws_bedrockagent_knowledge_base.test.arn}"
      }
    ]
}
EOF
}

resource "aws_iam_role" "kb" {
  name = "${var.rName}-kb"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "bedrock.amazonaws.com" }
      Action    = "sts:AssumeRole"
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = data.aws_caller_identity.current.account_id
        }
      }
    }]
  })
}

resource "aws_bedrockagent_knowledge_base" "test" {
  name     = var.rName
  role_arn = aws_iam_role.kb.arn

  knowledge_base_configuration {
    type = "MANAGED"

    managed_knowledge_base_configuration {
      embedding_model_type = "MANAGED"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
