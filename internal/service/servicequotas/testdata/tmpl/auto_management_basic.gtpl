resource "aws_servicequotas_auto_management" "test" {
{{- template "region" }}

  opt_in_level = "ACCOUNT"
  opt_in_type  = "NotifyOnly"
}