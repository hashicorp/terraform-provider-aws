resource "aws_devicefarm_project" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
