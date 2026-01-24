# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_parameter" "test" {
  count = 2

  name  = "${var.rName}-${count.index}"
  type  = "String"
  value = "${var.rName}-${count.index}"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
