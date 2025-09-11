# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_kms_alias" "test" {
  region = var.region

  name          = "alias/${var.rName}"
  target_key_id = aws_kms_key.test.id
}

resource "aws_kms_key" "test" {
  region = var.region

  description             = var.rName
  deletion_window_in_days = 7
  enable_key_rotation     = true
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
