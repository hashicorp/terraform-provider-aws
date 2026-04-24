# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_policy_attachment" "test" {
  count = var.resource_count

  name       = "${var.rName}-${count.index}"
  policy_arn = aws_iam_policy.test[count.index].arn
  groups     = [aws_iam_group.test[count.index].name]
  roles      = [aws_iam_role.test[count.index].name]
  users      = [aws_iam_user.test[count.index].name]
}

resource "aws_iam_policy" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "ec2:Describe*"
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_user" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
}

resource "aws_iam_group" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
}

resource "aws_iam_role" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
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

