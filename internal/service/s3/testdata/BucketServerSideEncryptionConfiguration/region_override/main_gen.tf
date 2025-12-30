# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  region = var.region

  bucket = aws_s3_bucket.test.bucket

  rule {
    # This is Amazon S3 bucket default encryption.
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
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
