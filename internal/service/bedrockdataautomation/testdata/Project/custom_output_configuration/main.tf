# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockdataautomation_blueprint" "test" {
  count = 2

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
}

resource "aws_bedrockdataautomation_project" "test" {
  name        = var.rName
  description = var.description

  custom_output_configuration {
    blueprint {
      blueprint_arn   = aws_bedrockdataautomation_blueprint.test[var.blueprint_index].arn
      blueprint_stage = "LIVE"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "description" {
  type     = string
  nullable = false
}

variable "blueprint_index" {
  type     = number
  nullable = false
}
