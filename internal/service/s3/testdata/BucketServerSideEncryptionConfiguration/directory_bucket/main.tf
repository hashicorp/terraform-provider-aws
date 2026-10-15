# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_directory_bucket.test.bucket

  rule {
    # This is Amazon S3 bucket default encryption.
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_directory_bucket" "test" {
  bucket = format(local.bucket_format, var.rName)

  location {
    name = local.location_name
  }
}

# testAccDirectoryBucketConfig_baseAZ

locals {
  location_name = data.aws_availability_zones.available.zone_ids[0]
  bucket_format = "%s--${local.location_name}--x-s3"
}

# testAccConfigDirectoryBucket_availableAZs

# acctest.ConfigAvailableAZsNoOptInExclude

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  exclude_zone_ids = ["use1-az1", "use1-az2", "use1-az3", "use2-az2", "usw2-az2", "aps1-az3", "apne1-az2", "euw1-az2"]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
