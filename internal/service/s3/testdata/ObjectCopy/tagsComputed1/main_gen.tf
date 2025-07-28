# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = var.rName
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"

  tagging_directive = "REPLACE"

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
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

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
