# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_insight" "test" {
  region = var.region

  filters {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = var.rName

  depends_on = [aws_securityhub_account.test]
}

resource "aws_securityhub_account" "test" {
  region = var.region

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
