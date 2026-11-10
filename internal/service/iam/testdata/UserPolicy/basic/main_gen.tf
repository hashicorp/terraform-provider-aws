# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_user_policy" "test" {
  name = var.rName
  user = aws_iam_user.test.name

  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_user" "test" {
  name = var.rName
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"
    actions = [
      "sts:GetCallerIdentity"
    ]
    resources = [
      "*"
    ]
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
