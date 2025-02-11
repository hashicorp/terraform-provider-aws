# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_m2_application" "test" {
  name        = var.rName
  engine_type = "bluage"
  definition {
    content = templatefile("test-fixtures/application-definition.json", { s3_bucket = aws_s3_bucket.test.id, version = "v1" })
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }

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

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
