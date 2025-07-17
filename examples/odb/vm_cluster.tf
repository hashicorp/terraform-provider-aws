//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

resource "aws_odb_cloud_vm_cluster" "vmcluster" {
  display_name                    = "Ofake_my_vmc"
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
  cpu_core_count                  = 6
  gi_version                      = "23.0.0.0"
  hostname_prefix                 = "apollo12"
  ssh_public_keys = [""] //public ssh keys
  odb_network_id                  = aws_odb_network.test.id
  is_local_backup_enabled         = true
  is_sparse_diskgroup_enabled     = true
  license_model                   = "LICENSE_INCLUDED"
  data_storage_size_in_tbs        = 20.0
  db_servers                      = [""] //db-servers
  db_node_storage_size_in_gbs     = 120.0
  memory_size_in_gbs              = 60
  tags = {
    "env" = "dev"
  }
}