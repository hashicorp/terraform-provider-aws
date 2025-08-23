# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

data "aws_availability_zones" "available" {
  # https://docs.aws.amazon.com/AmazonS3/latest/userguide/directory-bucket-az-networking.html#s3-express-endpoints-az.
  exclude_zone_ids = ["use1-az1", "use1-az2", "use1-az3", "use2-az2", "usw2-az2", "aps1-az3", "apne1-az2", "euw1-az2"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  location_name = data.aws_availability_zones.available.zone_ids[0]
  bucket        = "${var.rName}--${local.location_name}--x-s3"
}

resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }

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
