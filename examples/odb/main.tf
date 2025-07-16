//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

provider "aws" {
  region = "us-east-1"
}

resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name          = "Ofake-exa-4157426154220451194"
  shape             	= "Exadata.X9M"
  storage_count      	= 3
  compute_count         = 2
  availability_zone_id 	= "use1-az6"
  customer_contacts_to_send_to_oci = ["abc@example.com"]
  maintenance_window = {
    custom_action_timeout_in_mins = 16
    days_of_week =	[]
    hours_of_day =	[]
    is_custom_action_timeout_enabled = true
    lead_time_in_weeks = 0
    months = []
    patching_mode = "ROLLING"
    preference = "NO_PREFERENCE"
    weeks_of_month =[]
  }
}






resource "aws_odb_network" "test" {
  display_name          = "odb-net-6310376148776971562"
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access = "DISABLED"
  zero_etl_access = "DISABLED"
}



data "aws_odb_db_servers_list" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}

resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  cloud_exadata_infrastructure_id         = aws_odb_cloud_exadata_infrastructure.test.id
  odb_network_id                          =aws_odb_network.test.id
  display_name             				= "Ofake-avmc-1515523754357569237"
  autonomous_data_storage_size_in_tbs     = 5
  memory_per_oracle_compute_unit_in_gbs   = 2
  total_container_databases               = 1
  cpu_core_count_per_node                 = 40
  license_model                                = "LICENSE_INCLUDED"
  db_servers								   = [ for db_server in data.aws_odb_db_servers_list.test.db_servers : db_server.id]
  scan_listener_port_tls = 8561
  scan_listener_port_non_tls = 1024
  maintenance_window = {
    preference = "NO_PREFERENCE"
    days_of_week =	[]
    hours_of_day =	[]
    months = []
    weeks_of_month =[]
    lead_time_in_weeks = 0
  }

}


data "aws_odb_cloud_autonomous_vm_cluster" "test" {
  id             = aws_odb_cloud_autonomous_vm_cluster.test.id

}