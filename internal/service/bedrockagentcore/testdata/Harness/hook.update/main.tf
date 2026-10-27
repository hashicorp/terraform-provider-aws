# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = var.rName
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "Help me."
  }

  hook {
    before_invocation {
      name = "validate_request"

      target {
        lambda {
          arn = aws_lambda_function.test.arn
        }
      }
    }
  }

  hook {
    before_tool_call {
      name = "validate_tool_call"

      target {
        lambda {
          arn             = aws_lambda_function.test.arn
          timeout_seconds = 10
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_cloudwatch_event_bus" "test" {
  name = var.rName
}

resource "aws_sns_topic" "test" {
  name = var.rName
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = var.rName
  role          = aws_iam_role.lambda.arn
  handler       = "lambdatest.handler"
  runtime       = "nodejs24.x"
}

resource "aws_iam_role" "lambda" {
  name               = "${var.rName}-lambda"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

data "aws_iam_policy_document" "lambda_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
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
      "bedrock:InvokeModelWithResponseStream",
      "events:PutEvents",
      "lambda:InvokeFunction",
      "sns:Publish"
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
