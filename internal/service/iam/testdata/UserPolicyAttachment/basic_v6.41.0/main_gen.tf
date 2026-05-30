# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_user_policy_attachment" "test" {
  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_iam_user" "test" {
  name = var.rName
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
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.41.0"
    }
  }
}

provider "aws" {}
