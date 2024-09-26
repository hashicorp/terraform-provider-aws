resource "aws_sns_topic" "test" {
  name = var.rName

{{- template "tags" . }}
}
