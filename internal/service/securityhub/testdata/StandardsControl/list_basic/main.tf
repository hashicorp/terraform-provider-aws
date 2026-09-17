# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_standards_control" "test" {
  count = var.resource_count

  standards_control_arn = format("%s/1.1%d", replace(aws_securityhub_standards_subscription.test.id, "subscription", "control"), count.index)
  control_status        = "ENABLED"
}

resource "aws_securityhub_standards_subscription" "test" {
  standards_arn = "arn:${data.aws_partition.current.partition}:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
  depends_on    = [aws_securityhub_account.test]
}

resource "aws_securityhub_account" "test" {}

data "aws_partition" "current" {}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
