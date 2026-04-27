# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_acl" "test" {
  depends_on = [aws_s3_bucket_ownership_controls.test]

  bucket = aws_s3_bucket.test.bucket
  acl    = "private"
}

resource "aws_s3_bucket" "test" {
  bucket = var.rName
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
