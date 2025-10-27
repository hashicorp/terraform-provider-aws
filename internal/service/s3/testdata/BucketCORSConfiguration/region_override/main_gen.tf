# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_cors_configuration" "test" {
  region = var.region

  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
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
