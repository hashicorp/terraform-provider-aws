# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_group_policy_attachment" "test" {
  count = var.resource_count

  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test[count.index].arn
}

resource "aws_iam_group" "test" {
  name = var.rName
}

resource "aws_iam_policy" "test" {
  count = var.resource_count

  name   = "${var.rName}-${count.index}"
  policy = data.aws_iam_policy_document.test.json
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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
