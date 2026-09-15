# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ebs_volume_copy" "test" {
  region = var.region

  source_volume_id = aws_ebs_volume.test.id
}

resource "aws_ebs_volume" "test" {
  region = var.region

  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
  encrypted         = true
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  region = var.region

  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
