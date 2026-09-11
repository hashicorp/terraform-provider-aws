# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_odb_exascale_db_storage_vault" "test" {
  count = var.resource_count

  additional_flash_cache_in_percent                = 34
  autoscale_limit_in_gbs                           = 600
  availability_zone                                = data.aws_availability_zone.test.name
  availability_zone_id                             = local.availability_zone_id
  description                                      = "${var.rName}-${count.index} description"
  display_name                                     = "${var.rName}-${count.index}"
  high_capacity_database_storage_total_size_in_gbs = 300
  is_autoscale_enabled                             = true
  time_zone                                        = "UTC"

  tags = var.resource_tags
}

data "aws_availability_zone" "test" {
  zone_id = local.availability_zone_id
}

data "aws_region" "current" {}

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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource"
  type        = map(string)
  nullable    = false
}
