# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_user" "test" {
  name = var.rName

  tags = var.tags
}

variable "rName" {
  type     = string
  nullable = false
}

variable "tags" {
  type     = map(string)
  nullable = false
}


