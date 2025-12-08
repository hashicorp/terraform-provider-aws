# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_resourceexplorer2_view" "test" {
  region = var.region

  name = var.rName

  depends_on = [aws_resourceexplorer2_index.test]
}

resource "aws_resourceexplorer2_index" "test" {
  region = var.region

  type = "LOCAL"
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
