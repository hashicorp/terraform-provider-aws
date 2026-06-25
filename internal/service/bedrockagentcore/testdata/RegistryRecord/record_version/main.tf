# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_registry_record" "test" {
  name        = "${var.rName}-record"
  registry_id = aws_bedrockagentcore_registry.test.registry_id

  record_version = var.record_version

  descriptor_type = "CUSTOM"

  descriptors {
    custom {
      inline_content = "{}"
    }
  }
}

resource "aws_bedrockagentcore_registry" "test" {
  name = "${var.rName}-registry"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "record_version" {
  type     = string
  nullable = false
}