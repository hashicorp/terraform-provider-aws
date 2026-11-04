# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockdataautomation_blueprint" "test" {
  name = var.rName
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
        instruction   = var.instruction
      }
    }
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "instruction" {
  type     = string
  nullable = false
}
