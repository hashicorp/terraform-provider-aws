# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_agentregistry_registry" "test" {
  name = var.rName

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }

  approval_configuration {
    auto_approval_rules = ["APPROVE_ALL"]
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
