# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockdataautomation_blueprint" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
  type = "DOCUMENT"

  schema = jsonencode({
    "$schema"   = "http://json-schema.org/draft-07/schema#"
    description = "Pay stub"
    class       = "Pay stub"
    type        = "object"
    definitions = {}
    properties = {
      net_pay = {
        type          = "number"
        inferenceType = "explicit"
        instruction   = "The net pay for this pay period"
      }
    }
  })

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource"
  type        = map(string)
  nullable    = false
}
