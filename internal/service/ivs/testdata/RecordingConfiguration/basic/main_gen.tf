# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ivs_recording_configuration" "test" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
