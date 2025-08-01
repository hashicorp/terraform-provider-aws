//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

//Exadata Infrastructure with customer managed maintenance window
resource "aws_odb_cloud_exadata_infrastructure" "exa_infra_X11M_all_param" {
  display_name                        = "Ofake_my_odb_exadata_infra" //Required Field
  shape                               = "Exadata.X11M"  //Required Field
  storage_count                       = 3
  compute_count                       = 2
  availability_zone_id                = "use1-az6" //Required Field
  customer_contacts_to_send_to_oci    = ["abc@example.com"]
  database_server_type                = "X11M"
  storage_server_type                 = "X11M-HC"
  maintenance_window = { //Required
    custom_action_timeout_in_mins     = 16
    days_of_week                      = ["MONDAY", "TUESDAY"]
    hours_of_day                      = [11, 16]
    is_custom_action_timeout_enabled  = true
    lead_time_in_weeks                = 3
    months                            = ["FEBRUARY", "MAY", "AUGUST", "NOVEMBER"]
    patching_mode                     = "ROLLING"
    preference                        = "CUSTOM_PREFERENCE"
    weeks_of_month                    = [2, 4]
  }
  tags = {
    "env" = "dev"
  }

}

//Exadata Infrastructure with  default maintenance window with X9M system shape. with minimum parameters
resource "aws_odb_cloud_exadata_infrastructure" "test_X9M" {
  display_name          = "Ofake_my_exa_X9M"
  shape             	= "Exadata.X9M"
  storage_count      	= 3
  compute_count         = 2
  availability_zone_id 	= "use1-az6"
  maintenance_window = {
      custom_action_timeout_in_mins     = 16
      days_of_week                      = []
      hours_of_day                      = []
      is_custom_action_timeout_enabled  = true
      lead_time_in_weeks                = 0
      months                            = []
      patching_mode                     = "ROLLING"
      preference                        = "NO_PREFERENCE"
      weeks_of_month                    = []
    }
}