# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  count = var.resource_count

  bucket = aws_s3_bucket.test[count.index].bucket

  rule {
    # This is Amazon S3 bucket default encryption.
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
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
