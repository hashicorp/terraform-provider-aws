# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_user_policy_attachment" "test" {
  count = var.resource_count

  user       = aws_iam_user.test[count.index].name
  policy_arn = aws_iam_policy.test[count.index].arn
}

resource "aws_iam_user" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
}

resource "aws_iam_policy" "test" {
  count = var.resource_count

  name        = "${var.rName}-${count.index}"
  description = "A test policy"

  policy = data.aws_iam_policy_document.test[count.index].json
}

data "aws_iam_policy_document" "test" {
  count = var.resource_count

  statement {
    effect = "Allow"
    actions = [
      "iam:ChangePassword"
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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
