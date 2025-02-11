# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

resource "aws_m2_application" "test" {
  name        = var.rName
  engine_type = "bluage"
  definition {
    content = templatefile("test-fixtures/application-definition.json", { s3_bucket = aws_s3_bucket.test.id, version = "v1" })
  }

  tags = var.resource_tags

  depends_on = [aws_s3_object.test]
}

resource "aws_s3_bucket" "test" {
  bucket = var.rName
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "v1/PlanetsDemo-v1.zip"
  source = "test-fixtures/PlanetsDemo-v1.zip"
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
  nullable = false
}
