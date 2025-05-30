# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_resourceexplorer2_index" "test" {
  region = var.region

  type = "LOCAL"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
