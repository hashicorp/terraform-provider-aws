# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_account_policy" "test" {
  policy_name = var.rName
  policy_type = "FIELD_INDEX_POLICY"

  policy_document = jsonencode({
    Fields = ["RequestId"]
  })

  selection_criteria = "LogGroupNamePrefix = \"/${var.rName}\""
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.51.0"
    }
  }
}

provider "aws" {}
