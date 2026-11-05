# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = var.rName
  execution_role_arn = aws_iam_role.test.arn

  model {
    openai_model_config {
      api_key_arn = aws_bedrockagentcore_api_key_credential_provider.test.credential_provider_arn
      max_tokens  = 1000
      model_id    = "gpt-5"
      temperature = 0.95
      top_p       = 0.75

      additional_params = jsonencode({
        reasoning_effort = "high"
      })
    }
  }

  system_prompt {
    text = "Help me."
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_bedrockagentcore_api_key_credential_provider" "test" {
  name    = replace(var.rName, "_", "-")
  api_key = "test-api-key"
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
