# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_parameter" "test" {
  name  = var.rName
  type  = "String"
  value = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
