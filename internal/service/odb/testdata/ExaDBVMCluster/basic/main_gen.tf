# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_odb_network" "test" {
  availability_zone_id        = "use1-az6"
  backup_subnet_cidr          = "10.2.1.0/24"
  client_subnet_cidr          = "10.2.0.0/24"
  delete_associated_resources = true
  display_name                = "${var.rName}-network"
  s3_access                   = "DISABLED"
  zero_etl_access             = "DISABLED"
}

resource "aws_odb_exascale_db_storage_vault" "test" {
  availability_zone_id                             = "use1-az6"
  display_name                                     = "${var.rName}-vault"
  high_capacity_database_storage_total_size_in_gbs = 900
}

resource "aws_odb_exadb_vm_cluster" "test" {
  display_name                             = var.rName
  enabled_ecpu_count                       = 16
  exascale_db_storage_vault_id             = aws_odb_exascale_db_storage_vault.test.id
  grid_image_id                            = var.TF_AWS_ODB_EXADB_VM_CLUSTER_GRID_IMAGE_ID
  hostname                                 = "ofakevmc"
  node_count                               = 2
  odb_network_id                           = aws_odb_network.test.id
  shape                                    = "ExaDbXS"
  ssh_public_keys                          = [var.TF_AWS_ODB_EXADB_VM_CLUSTER_SSH_PUBLIC_KEY]
  total_ecpu_count                         = 64
  vm_file_system_storage_total_size_in_gbs = 440
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "TF_AWS_ODB_EXADB_VM_CLUSTER_GRID_IMAGE_ID" {
  type     = string
  nullable = false
}

variable "TF_AWS_ODB_EXADB_VM_CLUSTER_SSH_PUBLIC_KEY" {
  type     = string
  nullable = false
}
