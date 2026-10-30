# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockdataautomation_project" "test" {
  name = var.rName

  override_configuration {
    document {
      splitter {
        state = var.splitter_state
      }
    }

    modality_routing {
      png = "DOCUMENT"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "splitter_state" {
  type     = string
  nullable = false
}
