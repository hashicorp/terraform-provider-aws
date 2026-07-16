# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_workspaces_pool" "test" {
  bundle_id = data.aws_workspaces_bundle.standard.id
  capacity {
    desired_user_sessions = 1
  }
  description  = var.rName
  directory_id = aws_workspaces_directory.test.directory_id
  pool_name    = var.rName
  running_mode = "AUTO_STOP"
}

resource "aws_workspaces_directory" "test" {
  subnet_ids                      = [aws_subnet.primary.id, aws_subnet.secondary.id]
  workspace_type                  = "POOLS"
  workspace_directory_name        = var.rName
  workspace_directory_description = var.rName
  user_identity_type              = "CUSTOMER_MANAGED"
}

resource "aws_subnet" "primary" {
  vpc_id               = aws_vpc.test.id
  availability_zone_id = local.workspaces_az_ids[0]
  cidr_block           = "10.0.1.0/24"
}

resource "aws_subnet" "secondary" {
  vpc_id               = aws_vpc.test.id
  availability_zone_id = local.workspaces_az_ids[1]
  cidr_block           = "10.0.2.0/24"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

data "aws_workspaces_bundle" "standard" {
  owner = "AMAZON"
  name  = "Standard with Windows 10 (Server 2022 based) (WSP)"
}

data "aws_region" "current" {
}

data "aws_availability_zones" "available" {
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
