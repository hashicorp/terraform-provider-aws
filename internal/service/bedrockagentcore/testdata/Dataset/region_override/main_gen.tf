# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_dataset" "test" {
  region = var.region

  name        = var.rName
  schema_type = "AGENTCORE_EVALUATION_PREDEFINED_V1"

  source {
    inline_examples {
      examples = [
        jsonencode({
          scenario_id = "scenario-1"
          turns = [
            { input = "What is 2+2?", expected_response = "4" }
          ]
        })
      ]
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
