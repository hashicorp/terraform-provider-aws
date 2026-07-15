# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_policy_attachment" "test" {
  name       = var.rName
  groups     = aws_iam_group.test[*].name
  roles      = aws_iam_role.test[*].name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_iam_policy" "test" {
  name        = var.rName
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_group" "test" {
  count = 2
  name  = format("${var.rName}-%d", count.index + 1)
}

resource "aws_iam_role" "test" {
  count = 2
  name  = format("${var.rName}-%d", count.index + 1)

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
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
      version = "6.52.0"
    }
  }
}

provider "aws" {}
