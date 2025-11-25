# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_xray_group" "test" {
  group_name        = var.rName
  filter_expression = "responsetime > 5"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
