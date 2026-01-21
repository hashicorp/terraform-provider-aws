# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3tables_table_bucket_policy" "test" {
  region = var.region

  resource_policy  = data.aws_iam_policy_document.test.json
  table_bucket_arn = aws_s3tables_table_bucket.test.arn
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["s3tables:*"]
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
    resources = ["${aws_s3tables_table_bucket.test.arn}/*"]
  }
}

resource "aws_s3tables_table_bucket" "test" {
  region = var.region

  name = var.rName
}

data "aws_caller_identity" "current" {}

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
