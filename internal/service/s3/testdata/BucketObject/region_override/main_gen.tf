# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_object" "test" {
  region = var.region

  # Must have bucket versioning enabled first
  bucket = aws_s3_bucket_versioning.test.bucket
  key    = var.rName
}

resource "aws_s3_bucket" "test" {
  region = var.region

  bucket = var.rName
}

resource "aws_s3_bucket_versioning" "test" {
  region = var.region

  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
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
