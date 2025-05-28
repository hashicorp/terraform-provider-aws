# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_resourceexplorer2_index" "test" {
  type = "LOCAL"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
