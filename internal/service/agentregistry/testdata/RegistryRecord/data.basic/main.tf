# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_agentregistry_registry_record" "test" {
  registry_id = aws_agentregistry_registry.test.registry_id
  record_id   = aws_agentregistry_registry_record.test.record_id
}

resource "aws_agentregistry_registry" "test" {
  name = var.rName

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}

resource "aws_agentregistry_registry_record" "test" {
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
