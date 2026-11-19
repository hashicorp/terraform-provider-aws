# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_kms_key" "test" {
  count = 2

  description             = "${var.rName}-${count.index}"
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

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

  dynamic "encryption_configuration" {
    for_each = var.kms_key_index == null ? [] : [var.kms_key_index]

    content {
      kms_key_id = aws_kms_key.test[encryption_configuration.value].arn

      kms_encryption_context = {
        Purpose = var.context_value
      }
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "kms_key_index" {
  type     = number
  default  = null
  nullable = true
}

variable "context_value" {
  type     = string
  nullable = false
}
