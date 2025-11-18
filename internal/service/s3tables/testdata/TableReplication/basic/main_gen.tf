# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3tables_table_replication" "test" {
  table_arn = aws_s3tables_table.test.arn
}

resource "aws_s3tables_table" "test" {
  name             = var.rName
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"
}

resource "aws_s3tables_namespace" "test" {
  namespace        = var.rName
  table_bucket_arn = aws_s3tables_table_bucket.test.arn

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_s3tables_table_bucket" "test" {
  name = var.rName
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
