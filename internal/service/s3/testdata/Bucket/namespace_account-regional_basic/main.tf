# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket" "test" {
  bucket           = format("%s-%s-%s-an", var.rName, data.aws_caller_identity.current.account_id, data.aws_region.current.name)
  bucket_namespace = "account-regional"
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
