# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = "${var.rName}_s"
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = "USER_PREFERENCE"
  namespace_templates = ["/strategies/{memoryStrategyId}/actors/{actorId}/"]

  memory_record_schema {
    metadata_schema {
      key  = "ranking"
      type = "NUMBER"

      extraction_config {
        llm_extraction_config {
          definition = "Support ranking."

          validation {
            number_validation {
              min_value = 1.0
              max_value = 5.0
            }
          }
        }
      }
    }

    metadata_schema {
      key  = "vibes"
      type = "STRINGLIST"

      extraction_config {
        llm_extraction_config {
          definition = "Support vibes."

          validation {
            string_list_validation {
              allowed_values = ["fun", "spooky", "chill"]
              max_items      = 2
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
