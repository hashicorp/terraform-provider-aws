# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
  ignore_tags {
    keys = var.ignore_tag_keys
  }
}

resource "aws_bedrockagentcore_dataset" "test" {
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

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

variable "provider_tags" {
  type     = map(string)
  nullable = true
  default  = null
}

variable "ignore_tag_keys" {
  type     = set(string)
  nullable = false
}
