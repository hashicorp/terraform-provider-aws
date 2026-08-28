# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_odb_exascale_db_storage_vault" "test" {
  availability_zone_id                             = local.availability_zone_id
  display_name                                     = "ofake-${var.rName}"
  high_capacity_database_storage_total_size_in_gbs = 300

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
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

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
