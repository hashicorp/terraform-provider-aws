# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_insight" "test" {
  count = var.resource_count

  filters {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = "${var.rName}-${count.index}"

  depends_on = [aws_securityhub_account.test]
}

resource "aws_securityhub_account" "test" {}

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

