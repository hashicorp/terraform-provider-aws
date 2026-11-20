resource "aws_iot_thing_group" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
