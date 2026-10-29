# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = var.rName
  execution_role_arn = aws_iam_role.test.arn

  model {
    litellm_model_config {
      api_base    = "https://api.example.com/v1"
      model_id    = "anthropic/claude-sonnet-4-20250514"
      temperature = 0.7
      top_p       = 0.9

      additional_params = jsonencode({
        stop = ["END"]
      })
    }
  }

  system_prompt {
    text = "Help me."
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test.json
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

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": [
      "bedrock:InvokeModel",
      "bedrock:InvokeModelWithResponseStream"
    ],
    "Resource": "*"
  }
}
EOF
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
