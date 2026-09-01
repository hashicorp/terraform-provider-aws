# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_eip" "test" {
  region = var.region

  domain = "vpc"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
