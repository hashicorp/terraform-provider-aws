# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_policy" "test" {
  count = var.resource_count

  bucket = aws_s3_directory_bucket.test[count.index].bucket
  policy = data.aws_iam_policy_document.test[count.index].json
}

data "aws_iam_policy_document" "test" {
  count = var.resource_count

  statement {
    effect = "Allow"

    actions = [
      "s3express:*",
    ]

    resources = [
      aws_s3_directory_bucket.test[count.index].arn,
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}

resource "aws_s3_directory_bucket" "test" {
  count = var.resource_count

  bucket = format(local.bucket_format, var.rName, count.index)

  location {
    name = local.location_name
  }
}

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

# testAccDirectoryBucketConfig_baseAZ

locals {
  location_name = data.aws_availability_zones.available.zone_ids[0]
  bucket_format = "%s-%s--${local.location_name}--x-s3"
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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
