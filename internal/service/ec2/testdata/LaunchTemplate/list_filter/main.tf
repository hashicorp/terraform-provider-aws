# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_launch_template" "expected" {
  count = 2

  name = "${var.rName}-expected-${count.index}"
}

resource "aws_launch_template" "other" {
  name = "${var.rName}-other"
}

variable "rName" {
  type     = string
  nullable = false
}
