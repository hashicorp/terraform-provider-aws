# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0


resource "aws_odb_cloud_vm_cluster" "with_minimum_parameter" {
  display_name                    = "my-exa-infra"
  cloud_exadata_infrastructure_id = "exa_gjrmtxl4qk"
  cpu_core_count                  = 6
  gi_version                      = "23.0.0.0"
  hostname_prefix                 = "apollo12"
  ssh_public_keys                 = ["public-ssh-key"]
  odb_network_id                  = "odbnet_3l9st3litg"
  is_local_backup_enabled         = true
  is_sparse_diskgroup_enabled     = true
  license_model                   = "LICENSE_INCLUDED"
  data_storage_size_in_tbs        = 20.0
  db_servers                      = ["db-server-1", "db-server-2"]
  db_node_storage_size_in_gbs     = 120.0
  memory_size_in_gbs              = 60
  data_collection_options {
    is_diagnostics_events_enabled = false
    is_health_monitoring_enabled  = false
    is_incident_logs_enabled      = false
  }
}


resource "aws_odb_cloud_vm_cluster" "with_all_parameters" {
  display_name                    = "my-vmc"
  cloud_exadata_infrastructure_id = "exa_gjrmtxl4qk"
  cpu_core_count                  = 6
  gi_version                      = "23.0.0.0"
  hostname_prefix                 = "apollo12"
  ssh_public_keys                 = ["my-ssh-key"]
  odb_network_id                  = "odbnet_3l9st3litg"
  is_local_backup_enabled         = true
  is_sparse_diskgroup_enabled     = true
  license_model                   = "LICENSE_INCLUDED"
  data_storage_size_in_tbs        = 20.0
  db_servers                      = ["my-dbserver-1", "my-db-server-2"]
  db_node_storage_size_in_gbs     = 120.0
  memory_size_in_gbs              = 60
  cluster_name                    = "julia-13"
  timezone                        = "UTC"
  scan_listener_port_tcp          = 1521
  tags = {
    "env" = "dev"
  }
  data_collection_options {
    is_diagnostics_events_enabled = true
    is_health_monitoring_enabled  = true
    is_incident_logs_enabled      = true
  }
}