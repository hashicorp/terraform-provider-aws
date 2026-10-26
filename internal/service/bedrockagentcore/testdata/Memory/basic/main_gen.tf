# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_memory" "test" {
  name                  = var.rName
  event_expiry_duration = 7
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
