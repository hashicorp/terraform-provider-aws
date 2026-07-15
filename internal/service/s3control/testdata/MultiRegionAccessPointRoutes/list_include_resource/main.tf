# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3control_multi_region_access_point_routes" "test" {
  count = var.resource_count

  mrap = aws_s3control_multi_region_access_point.test[count.index].arn

  route {
    bucket                  = aws_s3_bucket.test_primary[count.index].bucket
    region                  = aws_s3_bucket.test_primary[count.index].bucket_region
    traffic_dial_percentage = 100
  }

  route {
    bucket                  = aws_s3_bucket.test_secondary[count.index].bucket
    region                  = aws_s3_bucket.test_secondary[count.index].bucket_region
    traffic_dial_percentage = 100
  }
}

resource "aws_s3_bucket" "test_primary" {
  count = var.resource_count

  bucket        = "${var.rName}-${count.index}-1"
  force_destroy = true
}

resource "aws_s3_bucket" "test_secondary" {
  count  = var.resource_count
  region = var.secondary_region

  bucket        = "${var.rName}-${count.index}-2"
  force_destroy = true
}

resource "aws_s3control_multi_region_access_point" "test" {
  count = var.resource_count

  details {
    name = "${var.rName}-${count.index}"

    region {
      bucket = aws_s3_bucket.test_primary[count.index].bucket
    }

    region {
      bucket = aws_s3_bucket.test_secondary[count.index].bucket
    }
  }
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

variable "secondary_region" {
  description = "Secondary region for MRAP buckets"
  type        = string
  nullable    = false
}
