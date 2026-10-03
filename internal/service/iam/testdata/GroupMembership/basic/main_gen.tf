# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_group_membership" "test" {
  name  = var.rName
  users = [aws_iam_user.test.name]
  group = aws_iam_group.test.name
}

resource "aws_iam_group" "test" {
  name = var.rName
}

resource "aws_iam_user" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
