# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_standards_control_association" "test" {
  region = var.region

  security_control_id = "IAM.1"
  standards_arn       = aws_securityhub_standards_subscription.test.standards_arn
  association_status  = "ENABLED"
}

data "aws_partition" "current" {}

resource "aws_securityhub_account" "test" {
  region = var.region

  enable_default_standards = false
}

resource "aws_securityhub_standards_subscription" "test" {
  region = var.region

  standards_arn = "arn:${data.aws_partition.current.partition}:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
  depends_on    = [aws_securityhub_account.test]
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
