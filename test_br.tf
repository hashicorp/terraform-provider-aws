# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}

provider "aws" {
  region = var.region
}

variable "region" {
  description = "AWS region for the test"
  type        = string
  default     = "us-east-1"
}

variable "policy_engine_name" {
  description = "Policy engine name (must start with a letter and use only letters, numbers, and underscores)"
  type        = string
  default     = "copilot_policy_engine_test"
}

variable "policy_engine_description" {
  description = "Policy engine description to test update behavior"
  type        = string
  default     = "initial description"
}

resource "aws_bedrockagentcore_policy_engine" "test" {
  name        = var.policy_engine_name
  description = var.policy_engine_description
}

output "policy_engine_id" {
  value = aws_bedrockagentcore_policy_engine.test.policy_engine_id
}

output "policy_engine_arn" {
  value = aws_bedrockagentcore_policy_engine.test.policy_engine_arn
}
