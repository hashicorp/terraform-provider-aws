resource "aws_apprunner_observability_configuration" "test" {
{{- template "region" }}
  observability_configuration_name = var.rName

{{- template "tags" . }}
}
