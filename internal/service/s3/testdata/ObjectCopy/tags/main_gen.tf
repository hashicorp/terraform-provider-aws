# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = var.rName
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"

  tagging_directive = "REPLACE"

  tags = var.resource_tags
}

resource "aws_s3_object" "source" {
  bucket  = aws_s3_bucket.source.bucket
  key     = var.rName
  content = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
}

resource "aws_s3_bucket" "source" {
  bucket = "${var.rName}-source"

  force_destroy = true
}

resource "aws_s3_bucket" "target" {
  bucket = "${var.rName}-target"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
