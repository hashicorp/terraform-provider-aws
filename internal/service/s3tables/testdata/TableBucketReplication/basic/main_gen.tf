# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3tables_table_bucket_replication" "test" {
  table_bucket_arn = aws_s3tables_table_bucket.test.arn
}

resource "aws_s3tables_table_bucket" "test" {
  name = var.rName
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
