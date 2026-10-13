# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_db_parameter_group" "test" {
  name   = var.rName
  family = "mysql5.6"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
