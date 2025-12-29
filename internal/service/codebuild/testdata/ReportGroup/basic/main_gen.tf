# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_codebuild_report_group" "test" {
  name = var.rName
  type = "TEST"

  export_config {
    type = "NO_EXPORT"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
