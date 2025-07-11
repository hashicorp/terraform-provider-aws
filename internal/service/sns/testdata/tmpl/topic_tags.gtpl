resource "aws_sns_topic" "test" {
  {{- template "region" . }}
  name = var.rName

{{- template "tags" . }}
}
