# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_memory_strategy" "test" {
  region = var.region

  name                = var.rName
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = "SEMANTIC"
  namespace_templates = ["default"]
}

resource "aws_bedrockagentcore_memory" "test" {
  region = var.region

  name                  = "${var.rName}_m"
  event_expiry_duration = 7
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
