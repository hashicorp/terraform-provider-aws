# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_policy" "test" {
  count = 3

  name = "${var.rName}-${count.index}"

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
