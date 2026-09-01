# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_metadata_configuration" "test" {
  region = var.region

  bucket = aws_s3_bucket.test.bucket

  metadata_configuration {
    inventory_table_configuration {
      configuration_state = "DISABLED"
    }

    journal_table_configuration {
      record_expiration {
        days       = 7
        expiration = "ENABLED"
      }
    }
  }
}

resource "aws_s3_bucket" "test" {
  region = var.region

  bucket = var.rName
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
