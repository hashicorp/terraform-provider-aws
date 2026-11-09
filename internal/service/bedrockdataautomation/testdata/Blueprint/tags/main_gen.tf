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

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
