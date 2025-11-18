# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3tables_table_replication" "test" {
  region = var.region

  table_arn = aws_s3tables_table.test.arn
}

resource "aws_s3tables_table" "test" {
  region = var.region

  name             = var.rName
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"
}

resource "aws_s3tables_namespace" "test" {
  region = var.region

  namespace        = var.rName
  table_bucket_arn = aws_s3tables_table_bucket.test.arn

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_s3tables_table_bucket" "test" {
  region = var.region

  name = var.rName
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
