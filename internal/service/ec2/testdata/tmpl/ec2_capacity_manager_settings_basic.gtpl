resource "aws_ec2_capacity_manager_settings" "test" {
{{- template "region" }}
  enabled = true
}
