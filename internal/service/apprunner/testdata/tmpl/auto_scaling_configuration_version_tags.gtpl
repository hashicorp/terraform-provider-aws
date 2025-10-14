resource "aws_apprunner_auto_scaling_configuration_version" "test" {
{{- template "region" }}
  auto_scaling_configuration_name = var.rName

{{- template "tags" . }}
}
