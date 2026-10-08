# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_policy" "test" {
  name             = var.rName
  policy_engine_id = aws_bedrockagentcore_policy_engine.test.policy_engine_id
  validation_mode  = "IGNORE_ALL_FINDINGS"

  definition {
    cedar {
      statement = var.statement
    }
  }
}

resource "aws_bedrockagentcore_policy_engine" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "statement" {
  type     = string
  nullable = false
}