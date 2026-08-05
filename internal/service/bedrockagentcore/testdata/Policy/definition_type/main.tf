# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_policy" "test" {
  name             = var.rName
  policy_engine_id = aws_bedrockagentcore_policy_engine.test.policy_engine_id
  validation_mode  = "IGNORE_ALL_FINDINGS"

  definition {
    dynamic "cedar" {
      for_each = var.variant == "cedar" ? [1] : []
      content {
        statement = var.statement
      }
    }
    dynamic "policy" {
      for_each = var.variant == "policy" ? [1] : []
      content {
        statement = var.statement
      }
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

variable "variant" {
  description = "Which definition union variant to set: cedar or policy"
  type        = string
  nullable    = false
}

variable "statement" {
  type     = string
  nullable = false
}
