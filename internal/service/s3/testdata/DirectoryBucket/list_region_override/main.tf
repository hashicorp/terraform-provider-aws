# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_directory_bucket" "test" {
  count  = var.resource_count
  region = var.region

  bucket = format(local.bucket_format, var.rName, count.index)

  location {
    name = local.location_name
  }
}

# testAccDirectoryBucketConfig_baseAZ

locals {
  location_name = data.aws_availability_zones.available.zone_ids[0]
  bucket_format = "%s-%s--${local.location_name}--x-s3"
}

# testAccConfigDirectoryBucket_availableAZs

locals {
  exclude_zone_ids = ["use1-az1", "use1-az2", "use1-az3", "use2-az2", "usw2-az2", "aps1-az3", "apne1-az2", "euw1-az2"]
}

# acctest.ConfigAvailableAZsNoOptInExclude

data "aws_availability_zones" "available" {
  region = var.region

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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
