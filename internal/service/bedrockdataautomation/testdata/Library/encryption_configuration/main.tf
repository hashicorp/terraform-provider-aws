# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_kms_key" "test" {
  count = 2

  description             = "${var.rName}-${count.index}"
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_bedrockdataautomation_library" "test" {
  name = var.rName

  encryption_configuration {
    kms_key_id = aws_kms_key.test[var.kms_key_index].arn

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

variable "kms_key_index" {
  type     = number
  nullable = false
}
