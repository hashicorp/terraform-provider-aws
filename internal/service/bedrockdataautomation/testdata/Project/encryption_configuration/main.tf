# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_kms_key" "test" {
  description             = var.rName
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_bedrockdataautomation_project" "test" {
  name = var.rName

  encryption_configuration {
    kms_key_id = aws_kms_key.test.arn

    kms_encryption_context = {
      Purpose = "test"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
