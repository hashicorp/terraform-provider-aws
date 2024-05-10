# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_user" "test" {
  name = var.rName

}

variable "rName" {
  type     = string
  nullable = false
}


