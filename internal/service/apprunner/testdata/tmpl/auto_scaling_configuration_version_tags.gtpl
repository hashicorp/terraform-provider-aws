resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = var.rName

{{- template "tags" . }}
}
