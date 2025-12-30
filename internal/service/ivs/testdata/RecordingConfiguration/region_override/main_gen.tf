# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ivs_recording_configuration" "test" {
  region = var.region

  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}

resource "aws_s3_bucket" "test" {
  region = var.region

  bucket        = var.rName
  force_destroy = true
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
