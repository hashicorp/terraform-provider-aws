resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = var.rName
{{- template "tags" . }}
}
