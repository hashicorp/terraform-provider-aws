# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_configuration_bundle" "test" {
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
