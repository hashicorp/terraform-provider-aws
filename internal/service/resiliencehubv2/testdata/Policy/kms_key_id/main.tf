# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehubv2_policy" "test" {
  name = var.rName

  availability_slo {
    target = 99.9
  }

  kms_key_id = aws_kms_key.test.arn
}

resource "aws_kms_key" "test" {
  description             = var.rName
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
