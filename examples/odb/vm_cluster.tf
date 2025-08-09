# Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

resource "aws_odb_cloud_vm_cluster" "my_vmcluster_with_minimum_parameters" {
  display_name                    = "Ofake_my_vmc"
  cloud_exadata_infrastructure_id = "<aws_odb_cloud_exadata_infrastructure.test.id>"
  cpu_core_count                  = 6
  gi_version                      = "23.0.0.0"
  hostname_prefix                 = "apollo12"
  ssh_public_keys                 = ["<public ssh keys>"]
  odb_network_id                  = "<aws_odb_network.test.id>"
  is_local_backup_enabled         = true
  is_sparse_diskgroup_enabled     = true
  license_model                   = "LICENSE_INCLUDED"
  data_storage_size_in_tbs        = 20.0
  db_servers                      = ["<my_deb_server>"] //db-servers
  db_node_storage_size_in_gbs     = 120.0
  memory_size_in_gbs              = 60
  tags = {
    "env" = "dev"
  }
}


resource "aws_odb_cloud_vm_cluster" "my_vmc_with_all_parameters" {
  display_name                        = "Ofake_my_vmc"
  cloud_exadata_infrastructure_id     = "<aws_odb_cloud_exadata_infrastructure.test.id>"
  cpu_core_count                      = 6
  gi_version                	        = "23.0.0.0"
  hostname_prefix                     = "apollo12"
  ssh_public_keys                     = ["<my_ssh_public_key>"]
  odb_network_id                      = "<aws_odb_network.test.id>"
  is_local_backup_enabled             = true
  is_sparse_diskgroup_enabled         = true
  license_model                       = "LICENSE_INCLUDED"
  data_storage_size_in_tbs            = 20.0
  db_servers					        = ["<my_db_server>"]
  db_node_storage_size_in_gbs         = 120.0
  memory_size_in_gbs                  = 60
  cluster_name              	      = "julia-13"
  timezone                            = "UTC"
  scan_listener_port_tcp		      = 1521
  system_version                      = "23.1.26.0.0.250516"
  tags = {
    "env" = "dev"
  }
  data_collection_options =  {
      is_diagnostics_events_enabled = true
      is_health_monitoring_enabled = true
      is_incident_logs_enabled = true
  }
}