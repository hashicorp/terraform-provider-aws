# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
  ignore_tags {
    keys = var.ignore_tag_keys
  }
}

resource "aws_odb_exascale_db_storage_vault" "test" {
  availability_zone_id                             = local.availability_zone_id
  display_name                                     = "ofake-${var.rName}"
  high_capacity_database_storage_total_size_in_gbs = 300

  tags = var.resource_tags
}

data "aws_region" "current" {
}

locals {
  availability_zone_ids = {
    "eu-west-1" = "euw1-az1"
    "us-east-1" = "use1-az2"
  }

  availability_zone_id = local.availability_zone_ids[data.aws_region.current.name]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

variable "provider_tags" {
  type     = map(string)
  nullable = true
  default  = null
}

variable "ignore_tag_keys" {
  type     = set(string)
  nullable = false
}
