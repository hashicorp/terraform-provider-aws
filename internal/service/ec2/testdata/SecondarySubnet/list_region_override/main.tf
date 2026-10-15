# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_secondary_subnet" "test" {
  count = var.resource_count

  region = var.region

  secondary_network_id = aws_ec2_secondary_network.test.id
  ipv4_cidr_block      = "10.0.${count.index}.0/24"
  availability_zone    = data.aws_availability_zones.available.names[0]
}

resource "aws_ec2_secondary_network" "test" {
  region = var.region

  ipv4_cidr_block = "10.0.0.0/16"
  network_type    = "rdma"
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
