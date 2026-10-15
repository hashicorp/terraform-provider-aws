# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3control_multi_region_access_point" "test" {
  details {
    name = var.rName

    region {
      bucket = aws_s3_bucket.test.id
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
