resource "aws_apprunner_observability_configuration" "test" {
  observability_configuration_name = var.rName

{{- template "tags" . }}
}
