# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_workspaces_pool" "test" {
  count  = var.resource_count
  region = var.region

  bundle_id = data.aws_workspaces_bundle.standard.id
  capacity {
    desired_user_sessions = 1
  }
  description  = "${var.rName}-${count.index}"
  directory_id = aws_workspaces_directory.test[count.index].directory_id
  pool_name    = "${var.rName}-${count.index}"
  running_mode = "AUTO_STOP"
}

resource "aws_workspaces_directory" "test" {
  count  = var.resource_count
  region = var.region

  subnet_ids                      = [aws_subnet.primary[count.index].id, aws_subnet.secondary[count.index].id]
  workspace_type                  = "POOLS"
  workspace_directory_name        = "${var.rName}-${count.index}"
  workspace_directory_description = "${var.rName}-${count.index}"
  user_identity_type              = "CUSTOMER_MANAGED"
}

resource "aws_subnet" "primary" {
  count  = var.resource_count
  region = var.region

  vpc_id               = aws_vpc.test.id
  availability_zone_id = local.workspaces_az_ids[0]
  cidr_block           = "10.0.${count.index * 2 + 1}.0/24"
}

resource "aws_subnet" "secondary" {
  count  = var.resource_count
  region = var.region

  vpc_id               = aws_vpc.test.id
  availability_zone_id = local.workspaces_az_ids[1]
  cidr_block           = "10.0.${count.index * 2 + 2}.0/24"
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.0.0.0/16"
}

data "aws_workspaces_bundle" "standard" {
  region = var.region

  owner = "AMAZON"
  name  = "Standard with Windows 10 (Server 2022 based) (WSP)"
}

data "aws_region" "current" {
  region = var.region
}

data "aws_availability_zones" "available" {
  region = var.region

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  region_workspaces_az_ids = {
    "us-east-1" = formatlist("use1-az%d", [2, 4, 6])
  }

  workspaces_az_ids = lookup(local.region_workspaces_az_ids, data.aws_region.current.name, data.aws_availability_zones.available.zone_ids)
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
