# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.1.0.0/16"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
