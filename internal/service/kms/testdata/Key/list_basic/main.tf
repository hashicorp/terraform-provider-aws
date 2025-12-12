# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_kms_key" "test" {
  count = 2

  description             = "${var.rName}-${count.index}"
  deletion_window_in_days = 7
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
