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

resource "aws_s3_bucket_object" "test" {
  # Must have bucket versioning enabled first
  bucket = aws_s3_bucket_versioning.test.bucket
  key    = var.rName

  tags = var.resource_tags
}

resource "aws_s3_bucket" "test" {
  bucket = var.rName
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.bucket
  versioning_configuration {
    status = "Enabled"
  }
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
