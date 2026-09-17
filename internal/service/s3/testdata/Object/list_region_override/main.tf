# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_object" "test" {
  count  = var.resource_count
  region = var.region

  bucket  = aws_s3_bucket.test.bucket
  key     = "${var.rName}-${count.index}"
  content = "test content"
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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
