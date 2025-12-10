# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_codeartifact_domain" "test" {
  region = var.region

  domain         = var.rName
  encryption_key = aws_kms_key.test.arn
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
