# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name =var.rName

  tags = {
    (var.tagKey1) = var.tagValue1
  }
}


variable "rName" {
  type     = string
  nullable = false
}

variable "tagKey1" {
  type     = string
  nullable = false
}

variable "tagValue1" {
  type     = string
  nullable = false
}


