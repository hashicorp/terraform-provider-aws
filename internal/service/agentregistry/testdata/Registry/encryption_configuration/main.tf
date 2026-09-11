# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_agentregistry_registry" "test" {
  name = var.rName

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }

  encryption_configuration {
    kms_key_arn = aws_kms_key.test.arn
  }
}

resource "aws_kms_key" "test" {
  description             = var.rName
  deletion_window_in_days = 7
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
