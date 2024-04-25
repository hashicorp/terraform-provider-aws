# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = var.rName

  tags = var.tags
}

variable "rName" {
  type     = string
  nullable = false
}

variable "tags" {
  type     = map(string)
  nullable = false
}


