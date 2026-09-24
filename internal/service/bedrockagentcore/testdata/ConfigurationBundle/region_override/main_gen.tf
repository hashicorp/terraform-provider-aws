# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_configuration_bundle" "test" {
  region = var.region

  bundle_name = var.rName

  component {
    component_identifier = "arn:aws:bedrock-agentcore:::evaluator/Builtin.Helpfulness"
    configuration        = jsonencode({ threshold = 0.7 })
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
