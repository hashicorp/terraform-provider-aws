# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3vectors_vector_bucket" "test" {
  vector_bucket_name = var.rName
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
