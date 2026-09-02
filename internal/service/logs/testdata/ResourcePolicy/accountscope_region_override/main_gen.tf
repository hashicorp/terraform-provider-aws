# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0


resource "aws_cloudwatch_log_resource_policy" "test" {
  region = var.region

  policy_name     = var.rName
  policy_document = data.aws_iam_policy_document.test.json
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/*"]

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
