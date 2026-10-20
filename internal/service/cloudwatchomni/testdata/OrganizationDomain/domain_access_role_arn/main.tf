# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatchomni_organization_domain" "test" {
  name                   = var.rName
  identity_providers     = ["IAM"]
  domain_access_role_arn = aws_iam_role.test[var.role_index].arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  count = 2

  name = "${var.rName}-${count.index}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "cloudwatch.amazonaws.com"
      }
      Action = ["sts:AssumeRole", "sts:TagSession"]
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = data.aws_caller_identity.current.account_id
        }
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  count = 2

  role       = aws_iam_role.test[count.index].name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/CloudWatchOmniDomainAccessPolicy"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "role_index" {
  description = "Index of the domain access role to use"
  type        = number
  nullable    = false
}
