resource "aws_ssm_service_setting" "test" {
{{- template "region" }}
  setting_id    = "/ssm/parameter-store/high-throughput-enabled"
  setting_value = "false"
}
