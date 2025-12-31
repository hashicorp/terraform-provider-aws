# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_policy" "expected" {
  count = 2

  name = "${var.rName}-${count.index}"
  path = var.expected_path_name

  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_policy" "not_expected" {
  count = 2

  name = "not-${var.rName}-${count.index}"
  path = var.other_path_name

  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"
    actions = [
      "ec2:Describe*"
    ]
    resources = ["*"]
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "expected_path_name" {
  description = "Path name for expected resources"
  type        = string
  nullable    = false
}

variable "other_path_name" {
  description = "Path name for non-expected resources"
  type        = string
  nullable    = false
}
