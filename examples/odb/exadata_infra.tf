# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

# Exadata Infrastructure with customer managed maintenance window
resource "aws_odb_cloud_exadata_infrastructure" "exa_infra_all_params" {
  display_name                     = "Ofake-my-exa-infra"
  shape                            = "Exadata.X11M"
  storage_count                    = 3
  compute_count                    = 2
  availability_zone_id             = "use1-az6"
  customer_contacts_to_send_to_oci = [{ email = "abc@example.com" }, { email = "def@example.com" }]
  database_server_type             = "X11M"
  storage_server_type              = "X11M-HC"
  maintenance_window {
    custom_action_timeout_in_mins    = 16
    days_of_week                     = [{ name = "MONDAY" }, { name = "TUESDAY" }]
    hours_of_day                     = [11, 16]
    is_custom_action_timeout_enabled = true
    lead_time_in_weeks               = 3
    months                           = [{ name = "FEBRUARY" }, { name = "MAY" }, { name = "AUGUST" }, { name = "NOVEMBER" }]
    patching_mode                    = "ROLLING"
    preference                       = "CUSTOM_PREFERENCE"
    weeks_of_month                   = [2, 4]
  }
  tags = {
    "env" = "dev"
  }

}

# Exadata Infrastructure with  default maintenance window with X9M system shape. with minimum parameters
resource "aws_odb_cloud_exadata_infrastructure" "exa_infra_basic" {
  display_name         = "Ofake_my_exa_X9M"
  shape                = "Exadata.X9M"
  storage_count        = 3
  compute_count        = 2
  availability_zone_id = "use1-az6"
  maintenance_window {
    custom_action_timeout_in_mins    = 16
    is_custom_action_timeout_enabled = true
    patching_mode                    = "ROLLING"
    preference                       = "NO_PREFERENCE"
  }
}
