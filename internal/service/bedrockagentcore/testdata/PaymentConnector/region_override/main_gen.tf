# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_payment_connector" "test" {
  region = var.region

  name               = var.rName
  payment_manager_id = aws_bedrockagentcore_payment_manager.test.payment_manager_id
  type               = "StripePrivy"

  credential_provider_configuration {
    stripe_privy {
      credential_provider_arn = aws_bedrockagentcore_payment_credential_provider.test.credential_provider_arn
    }
  }
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

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "bedrock-agentcore:CreateWorkloadIdentity",
          "bedrock-agentcore:TagResource",
          "bedrock-agentcore:UntagResource",
          "bedrock-agentcore:ListTagsForResource",
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_bedrockagentcore_payment_manager" "test" {
  region = var.region

  name            = replace(replace(var.rName, "-", ""), "_", "")
  authorizer_type = "AWS_IAM"
  role_arn        = aws_iam_role.test.arn
}

resource "aws_bedrockagentcore_payment_credential_provider" "test" {
  region = var.region

  name                       = replace(replace(var.rName, "-", ""), "_", "")
  credential_provider_vendor = "StripePrivy"

  provider_configuration {
    stripe_privy_configuration {
      app_id                    = "app_test_id"
      app_secret                = "sk_test_secret"
      authorization_id          = "auth_test_id"
      authorization_private_key = "MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgNlzLhFfPe/14eR6GlFuWOzYTHgfXgyKs1yHwtpFISo6hRANCAAQPrRtegKcGCGBALTzewz0OnIpa9AeOe5BpcT0OS+Ej7odZ7fsTN8YgZzq5kBAY3u2UcZNHn6YJC70Z4bgpiuKI"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
