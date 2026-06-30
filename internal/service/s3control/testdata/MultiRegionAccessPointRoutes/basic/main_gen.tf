# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3control_multi_region_access_point_routes" "test" {
  mrap = aws_s3control_multi_region_access_point.test.arn

  route {
    bucket                  = aws_s3_bucket.test1.bucket
    region                  = aws_s3_bucket.test1.bucket_region
    traffic_dial_percentage = 100
  }

  route {
    bucket                  = aws_s3_bucket.test2.bucket
    region                  = aws_s3_bucket.test2.bucket_region
    traffic_dial_percentage = 100
  }
}

resource "aws_s3_bucket" "test1" {
  bucket        = "${var.rName}-1"
  force_destroy = true
}

resource "aws_s3_bucket" "test2" {
  region = var.secondary_region

  bucket        = "${var.rName}-2"
  force_destroy = true
}

resource "aws_s3control_multi_region_access_point" "test" {
  details {
    name = var.rName

    region {
      bucket = aws_s3_bucket.test1.bucket
    }

    region {
      bucket = aws_s3_bucket.test2.bucket
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "secondary_region" {
  description = "Secondary region"
  type        = string
  nullable    = false
}
