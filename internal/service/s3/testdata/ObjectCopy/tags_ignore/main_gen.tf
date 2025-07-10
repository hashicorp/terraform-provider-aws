# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
  ignore_tags {
    keys = var.ignore_tag_keys
  }
}

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

variable "provider_tags" {
  type     = map(string)
  nullable = true
  default  = null
}

variable "ignore_tag_keys" {
  type     = set(string)
  nullable = false
}
