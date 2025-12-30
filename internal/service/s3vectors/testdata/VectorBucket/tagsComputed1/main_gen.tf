# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_s3vectors_vector_bucket" "test" {
  vector_bucket_name = var.rName

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
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
