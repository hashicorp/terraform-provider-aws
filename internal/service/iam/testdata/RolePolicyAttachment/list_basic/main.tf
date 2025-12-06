# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_role_policy_attachment" "customer_managed" {
  count = 2

  role       = aws_iam_role.test[count.index].name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_iam_role_policy_attachment" "aws_managed" {
  count = 2

  role       = aws_iam_role.test[count.index].name
  policy_arn = data.aws_iam_policy.AmazonDynamoDBReadOnlyAccess.arn
}

resource "aws_iam_role" "test" {
  count = 2

  name = "${var.rName}-${count.index}"

  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = ["ec2.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_policy" "test" {
  name        = var.rName
  description = "A test policy"

  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
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

data "aws_iam_policy" "AmazonDynamoDBReadOnlyAccess" {
  name = "AmazonDynamoDBReadOnlyAccess"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
