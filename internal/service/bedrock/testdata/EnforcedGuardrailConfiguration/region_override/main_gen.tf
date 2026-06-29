# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrock_guardrail" "test" {
  region = var.region

  name                      = var.rName
  blocked_input_messaging   = "Blocked input"
  blocked_outputs_messaging = "Blocked output"
  description               = "Test guardrail for enforced guardrail configuration"

  word_policy_config {
    words_config {
      text = "deny"
    }
  }
}

resource "aws_bedrock_guardrail_version" "test" {
  region = var.region

  guardrail_arn = aws_bedrock_guardrail.test.guardrail_arn
  description   = "Test guardrail version"
}

resource "aws_bedrock_enforced_guardrail_configuration" "test" {
  region = var.region

  guardrail_identifier = aws_bedrock_guardrail.test.guardrail_arn
  guardrail_version    = aws_bedrock_guardrail_version.test.version
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
