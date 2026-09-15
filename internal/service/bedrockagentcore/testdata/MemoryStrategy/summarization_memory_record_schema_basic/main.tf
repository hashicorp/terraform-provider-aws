# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = "${var.rName}_s"
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = "SUMMARIZATION"
  namespace_templates = ["/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"]

  memory_record_schema {
    metadata_schema {
      key = var.key
    }
  }
}

resource "aws_bedrockagentcore_memory" "test" {
  name                  = "${var.rName}_m"
  event_expiry_duration = 7

  indexed_key {
    key  = "customer_id"
    type = "STRING"
  }

  indexed_key {
    key  = "score"
    type = "NUMBER"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "key" {
  type     = string
  nullable = false
  validation {
    condition     = contains(["customer_id", "score"], var.key)
    error_message = "key_name must be one of: customer_id, score."
  }
}