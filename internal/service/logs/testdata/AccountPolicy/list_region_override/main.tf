# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_account_policy" "test" {
  count = var.resource_count

  region = var.region

  policy_name = "${var.rName}-${count.index}"
  policy_type = "RETENTION_POLICY"

  policy_document = jsonencode({
    retentionInDays = 90
    enforcementMode = "NEW_ONLY"
  })

  selection_criteria = "LogGroupNamePrefix = \"/${var.rName}-${count.index}\""
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
