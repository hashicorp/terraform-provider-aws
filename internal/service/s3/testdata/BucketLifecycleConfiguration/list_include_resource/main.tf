# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  count = var.resource_count

  bucket = aws_s3_bucket.test[count.index].bucket
  rule {
    id     = var.rName
    status = "Enabled"

    expiration {
      days = 365
    }

    filter {
      prefix = "prefix/"
    }
  }
}

resource "aws_s3_bucket" "test" {
  count = var.resource_count

  bucket = "${var.rName}-${count.index}"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
