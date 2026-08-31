# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_odb_network" "test" {
  region = var.region

  availability_zone_id        = local.availability_zone_id
  backup_subnet_cidr          = "10.2.1.0/24"
  client_subnet_cidr          = "10.2.0.0/24"
  delete_associated_resources = true
  display_name                = "${var.rName}-network"
  s3_access                   = "DISABLED"
  zero_etl_access             = "DISABLED"
}

resource "aws_odb_exascale_db_storage_vault" "test" {
  region = var.region

  availability_zone_id                             = local.availability_zone_id
  display_name                                     = "${var.rName}-vault"
  high_capacity_database_storage_total_size_in_gbs = 900
}

resource "aws_odb_exadb_vm_cluster" "test" {
  count  = var.resource_count
  region = var.region

  display_name                             = "${var.rName}-${count.index}"
  enabled_ecpu_count                       = 16
  exascale_db_storage_vault_id             = aws_odb_exascale_db_storage_vault.test.id
  grid_image_id                            = var.grid_image_id
  hostname                                 = "ofake${count.index}${var.hostname_suffix}"
  node_count                               = 2
  odb_network_id                           = aws_odb_network.test.id
  shape                                    = "ExaDbXS"
  ssh_public_keys                          = [var.ssh_public_key]
  total_ecpu_count                         = 64
  vm_file_system_storage_total_size_in_gbs = 440
}

data "aws_region" "current" {
  region = var.region
}

locals {
  availability_zone_ids = {
    "eu-west-1" = "euw1-az3"
    "us-east-1" = "use1-az6"
  }

  availability_zone_id = local.availability_zone_ids[data.aws_region.current.name]
}

variable "grid_image_id" {
  description = "Grid Infrastructure image ID"
  type        = string
  nullable    = false
}

variable "hostname_suffix" {
  description = "Unique hostname suffix"
  type        = string
  nullable    = false
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "ssh_public_key" {
  description = "SSH public key"
  type        = string
  nullable    = false
}
