# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpc_encryption_control" "test" {
  region = var.region

  vpc_id = aws_vpc.test.id
  mode   = "monitor"
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.1.0.0/16"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
