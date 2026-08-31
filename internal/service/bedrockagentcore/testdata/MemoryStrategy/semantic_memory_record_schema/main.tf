# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = "${var.rName}_s"
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = "SEMANTIC"
  namespace_templates = ["/strategies/{memoryStrategyId}/actors/{actorId}/"]

  memory_record_schema {
    metadata_schema {
      extraction_type = "LLM_INFERRED"
      key             = "priority"
      type            = "STRING"

      extraction_config {
        llm_extraction_config {
          definition = var.definition

          validation {
            string_validation {
              allowed_values = ["critical", "high", "medium", "low"]
            }
          }
        }
      }
    }
  }
}

resource "aws_bedrockagentcore_memory" "test" {
  name                  = "${var.rName}_m"
  event_expiry_duration = 7
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "definition" {
  type     = string
  nullable = false
}