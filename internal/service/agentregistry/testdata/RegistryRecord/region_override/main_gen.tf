# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_agentregistry_registry" "test" {
  region = var.region

  name = var.rName
  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}

resource "aws_agentregistry_registry_record" "test" {
  region = var.region

  registry_id = aws_agentregistry_registry.test.registry_id
  name        = var.rName
  record_type = "CUSTOM"

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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
