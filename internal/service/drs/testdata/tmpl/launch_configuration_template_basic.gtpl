resource "aws_drs_launch_configuration_template" "test" {
  copy_private_ip             = true
  copy_tags                   = false
  launch_disposition          = "STARTED"
  launch_into_source_instance = false
  post_launch_enabled         = false
  recovery_mode               = "OPTIMAL"

  licensing {
    os_byol = true
  }

  target_instance_type_right_sizing_method = "IN_AWS"
{{- template "tags" . }}
}
