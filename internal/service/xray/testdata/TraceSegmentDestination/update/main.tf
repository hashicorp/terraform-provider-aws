# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_xray_trace_segment_destination" "test" {
  destination = var.destination

  depends_on = [aws_cloudwatch_log_resource_policy.test]
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:PutLogEvents",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:logs:*:*:log-group:aws/spans:*",
      "arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/application-signals/data:*"
    ]

    principals {
      identifiers = ["xray.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = var.rName
  policy_document = data.aws_iam_policy_document.test.json
}

variable "destination" {
  description = "Destination"
  type        = string
  nullable    = false
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}