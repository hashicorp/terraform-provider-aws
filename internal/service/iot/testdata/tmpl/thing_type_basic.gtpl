resource "aws_iot_thing_type" "test" {
{{- template "region" }}
  name = var.rName
{{- template "tags" . }}
}
