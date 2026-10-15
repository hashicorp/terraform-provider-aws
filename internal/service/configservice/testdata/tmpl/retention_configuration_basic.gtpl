resource "aws_config_retention_configuration" "test" {
{{- template "region" }}
  retention_period_in_days = 90
}
