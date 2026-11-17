# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_cors_configuration" "test" {
  count = var.resource_count

  bucket = aws_s3_bucket.test[count.index].bucket

  cors_rule {
    allowed_methods = ["GET"]
    allowed_origins = ["https://example.com"]
  }
}

resource "aws_s3_bucket" "test" {
  count = var.resource_count

  bucket = "${var.rName}-${count.index}"
}

resource "aws_s3_bucket" "no_cors" {
  bucket = "${var.rName}-no-cors"
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
