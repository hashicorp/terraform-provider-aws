resource "aws_workspaces_pool" "test" {
{{- template "region" }}
  bundle_id = data.aws_workspaces_bundle.standard.id
  capacity {
    desired_user_sessions = 1
  }
  description  = var.rName
  directory_id = aws_workspaces_directory.test.directory_id
  pool_name    = var.rName
  running_mode = "AUTO_STOP"
{{- template "tags" . }}
}

resource "aws_workspaces_directory" "test" {
{{- template "region" }}
  subnet_ids                      = [aws_subnet.primary.id, aws_subnet.secondary.id]
  workspace_type                  = "POOLS"
  workspace_directory_name        = var.rName
  workspace_directory_description = var.rName
  user_identity_type              = "CUSTOMER_MANAGED"
}

resource "aws_subnet" "primary" {
{{- template "region" }}
  vpc_id               = aws_vpc.test.id
  availability_zone_id = local.workspaces_az_ids[0]
  cidr_block           = "10.0.1.0/24"
}

resource "aws_subnet" "secondary" {
{{- template "region" }}
  vpc_id               = aws_vpc.test.id
  availability_zone_id = local.workspaces_az_ids[1]
  cidr_block           = "10.0.2.0/24"
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"
}

data "aws_workspaces_bundle" "standard" {
{{- template "region" }}
  owner = "AMAZON"
  name  = "Standard with Windows 10 (Server 2022 based) (WSP)"
}

data "aws_region" "current" {
{{- template "region" }}
}

data "aws_availability_zones" "available" {
{{- template "region" }}
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
