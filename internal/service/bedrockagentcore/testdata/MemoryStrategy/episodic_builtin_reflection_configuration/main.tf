# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = "${var.rName}_s"
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = "EPISODIC"
  namespace_templates = [var.namespace_template]

  reflection_configuration {
    namespace_templates = [var.reflection_namespace_template]
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

variable "namespace_template" {
  type     = string
  nullable = false
}

variable "reflection_namespace_template" {
  type     = string
  nullable = false
}