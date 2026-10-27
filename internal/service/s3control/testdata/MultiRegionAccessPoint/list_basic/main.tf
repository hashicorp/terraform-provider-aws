# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3control_multi_region_access_point" "test" {
  count = var.resource_count

  details {
    name = "${var.rName}-${count.index}"

    region {
      bucket = aws_s3_bucket.test[count.index].id
    }
  }
}

resource "aws_s3_bucket" "test" {
  count = var.resource_count

  bucket        = "${var.rName}-${count.index}"
  force_destroy = true
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
