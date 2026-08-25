# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_memory" "test" {
  name                  = var.rName
  event_expiry_duration = 7

  dynamic "indexed_key" {
    for_each = var.keys

    content {
      key  = indexed_key.value
      type = local.indexed_keys[indexed_key.value]
    }
  }
}

locals {
  indexed_keys = {
    "customer_id" = "STRING"
    "priority"    = "STRING"
    "score"       = "NUMBER"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "keys" {
  description = "List of keys to be indexed"
  type        = list(string)
  nullable    = false
}