# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket" "test" {
  bucket_prefix    = var.rName
  bucket_namespace = "account-regional"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
