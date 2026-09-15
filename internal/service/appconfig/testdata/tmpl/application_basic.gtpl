resource "aws_appconfig_application" "test" {
  name = var.rName

{{- template "tags" . }}
}
