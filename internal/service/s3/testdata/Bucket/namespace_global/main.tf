# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket" "test" {
  bucket           = var.rName
  bucket_namespace = "global"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
