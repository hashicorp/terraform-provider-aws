# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_agentregistry_registry" "test" {
  name = var.rName

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}

resource "aws_agentregistry_registry_record" "test" {
  registry_id  = aws_agentregistry_registry.test.registry_id
  name         = var.rName
  record_type  = "CUSTOM"
  description  = var.description
  display_name = var.display_name

  descriptors {
    custom {
      data = jsonencode({
        name = var.rName
      })
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

variable "display_name" {
  type     = string
  nullable = false
}
