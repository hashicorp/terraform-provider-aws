# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

# testAccDirectoryBucketConfig_baseAZ

locals {
  location_name = data.aws_availability_zones.available.zone_ids[0]
  bucket        = "${var.rName}--${local.location_name}--x-s3"
}

# testAccConfigDirectoryBucket_availableAZs

locals {
  # https://docs.aws.amazon.com/AmazonS3/latest/userguide/directory-bucket-az-networking.html#s3-express-endpoints-az.
  exclude_zone_ids = ["use1-az1", "use1-az2", "use1-az3", "use2-az2", "usw2-az2", "aps1-az3", "apne1-az2", "euw1-az2"]
}

# acctest.ConfigAvailableAZsNoOptInExclude

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.31.0"
    }
  }
}

provider "aws" {}
