# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_object" "test" {
  count = var.resource_count

  bucket  = aws_s3_bucket.test.bucket
  key     = "${var.rName}-${count.index}"
  content = "test content"

  tags = var.resource_tags
}

resource "aws_s3_bucket" "test" {
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

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
