# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = var.rName

}

variable "rName" {
  type     = string
  nullable = false
}


