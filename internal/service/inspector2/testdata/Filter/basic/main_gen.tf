# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_inspector2_filter" "test" {
  name   = var.rName
  action = "NONE"
  filter_criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "111222333444"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
