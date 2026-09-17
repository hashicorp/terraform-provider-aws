# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
  ignore_tags {
    keys = var.ignore_tag_keys
  }
}

resource "aws_dataexchange_revision_assets" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  asset {
    import_assets_from_s3 {
      asset_source {
        bucket = aws_s3_object.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }

  tags = var.resource_tags
}

resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = var.rName
  name        = var.rName
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "test"
  content = "test"
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
