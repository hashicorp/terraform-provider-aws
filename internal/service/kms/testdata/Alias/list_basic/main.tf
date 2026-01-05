# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_kms_key" "test" {
  count = 2

  description             = "${var.rName}-${count.index}"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  count = 2

  name          = "alias/${var.rName}-${count.index}"
  target_key_id = aws_kms_key.test[count.index].key_id
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
