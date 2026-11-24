# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockdataautomation_library" "test" {
  name        = var.rName
  description = var.description
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "description" {
  type     = string
  default  = null
  nullable = true
}
