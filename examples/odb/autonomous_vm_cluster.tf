# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

# Autonomous VM Cluster with default maintenance window and minimum parameters
resource "aws_odb_cloud_autonomous_vm_cluster" "avmc_with_minimum_parameters" {
  cloud_exadata_infrastructure_id       = "<exadata_infra_id>" # refer your exadata infra id
  odb_network_id                        = "<odb_net_id>"       # refer_your_odb_net_id
  display_name                          = "Ofake-avmc-my_avmc"
  autonomous_data_storage_size_in_tbs   = 5
  memory_per_oracle_compute_unit_in_gbs = 2
  total_container_databases             = 1
  cpu_core_count_per_node               = 40
  license_model                         = "LICENSE_INCLUDED"
  # ids of db server. refer your exa infra. This is a manadatory fileld. Refer your cloud exadata infrastructure for db server id
  db_servers                 = ["<my_db_server_id>"]
  scan_listener_port_tls     = 8561
  scan_listener_port_non_tls = 1024
  maintenance_window {
    preference = "NO_PREFERENCE"
  }

}

# Autonomous VM Cluster with all parameters
resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  description                           = "my first avmc"
  time_zone                             = "UTC"
  cloud_exadata_infrastructure_id       = "<aws_odb_cloud_exadata_infrastructure.test.id>"
  odb_network_id                        = "<aws_odb_network.test.id>"
  display_name                          = "Ofake_my avmc"
  autonomous_data_storage_size_in_tbs   = 5
  memory_per_oracle_compute_unit_in_gbs = 2
  total_container_databases             = 1
  cpu_core_count_per_node               = 40
  license_model                         = "LICENSE_INCLUDED"
  db_servers                            = ["<my_db_server_1>", "<my_db_server_2>"]
  scan_listener_port_tls                = 8561
  scan_listener_port_non_tls            = 1024
  maintenance_window {
    days_of_week       = [{ name = "MONDAY" }, { name = "TUESDAY" }]
    hours_of_day       = [4, 16]
    lead_time_in_weeks = 3
    months             = [{ name = "FEBRUARY" }, { name = "MAY" }, { name = "AUGUST" }, { name = "NOVEMBER" }]
    preference         = "CUSTOM_PREFERENCE"
    weeks_of_month     = [2, 4]
  }
  tags = {
    "env" = "dev"
  }

}
