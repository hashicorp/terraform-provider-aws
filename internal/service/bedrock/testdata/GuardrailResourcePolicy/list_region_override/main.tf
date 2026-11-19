# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrock_guardrail_resource_policy" "test" {
  count  = var.resource_count
  region = var.region

  resource_arn = aws_bedrock_guardrail.test[count.index].guardrail_arn

  policy = data.aws_iam_policy_document.test[count.index].json
}

data "aws_iam_policy_document" "test" {
  count = var.resource_count

  statement {
    effect    = "Allow"
    actions   = ["bedrock:ApplyGuardrail", "bedrock:GetGuardrail"]
    resources = [aws_bedrock_guardrail.test[count.index].guardrail_arn]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_bedrock_guardrail" "test" {
  count  = var.resource_count
  region = var.region

  name                      = "${var.rName}-${count.index}"
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"

  content_policy_config {
    filters_config {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "HATE"
    }
  }
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
