# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_xray_encryption_config" "test" {
  region = var.region

  type = "NONE"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
