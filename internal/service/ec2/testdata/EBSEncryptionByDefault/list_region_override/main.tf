# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ebs_encryption_by_default" "test" {
  enabled = true
  region  = var.region
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
