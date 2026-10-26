# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_index_policy" "test" {
  region = var.region

  log_group_name  = aws_cloudwatch_log_group.test.name
  policy_document = "{\"Fields\":[\"eventName\"]}"
}

resource "aws_cloudwatch_log_group" "test" {
  region = var.region

  name = "/aws/testacc/index-policy-${var.rName}"
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
