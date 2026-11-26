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

data "aws_odb_gi_minor_versions" "test" {
  region = var.region

  availability_zone_id = local.availability_zone_id
  gi_version           = "26.0.0.0"
  shape_family         = "EXADB_XS"
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
  test_ssh_public_key  = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDNt3kA/dBkS6ZyU/sVDiGMuWJQaRPmLNbs/25K/e/fIl07ZWUgqqsFkcycLLMNFGD30Cmgp6XCXfNlIjzFWhNam+4cBb4DPpvieUw44VgsHK5JQy3JKlUfglmH5rs4G5pLiVfZpFU6jqvTsu4mE1CHCP0sXJlJhGxMG3QbsqYWNKiqGFEhuzGMs6fQlMkNiXsFoDmh33HAcXCbaFSC7V7xIqT1hlKu0iOL+GNjMj4R3xy0o3jafhO4MG2s3TwCQQCyaa5oyjL8iP8p3L9yp6cbIcXaS72SIgbCSGCyrcQPIKP2lJJHvE1oVWzLVBhR4eSzrlFDv7K4IErzaJmHqdiz" # nosemgrep:ci.ssh-key
}

resource "aws_odb_exadb_vm_cluster" "test" {
  region = var.region

  display_name                             = var.rName
  enabled_ecpu_count                       = 16
  exascale_db_storage_vault_id             = aws_odb_exascale_db_storage_vault.test.id
  grid_image_id                            = data.aws_odb_gi_minor_versions.test.gi_minor_versions[0].grid_image_id
  hostname                                 = "ofakevmc"
  node_count                               = 2
  odb_network_id                           = aws_odb_network.test.id
  shape                                    = "EXADBXS"
  ssh_public_keys                          = [local.test_ssh_public_key]
  total_ecpu_count                         = 64
  vm_file_system_storage_total_size_in_gbs = 440
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
