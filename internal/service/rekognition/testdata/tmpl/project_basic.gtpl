resource "aws_rekognition_project" "test" {
{{- template "region" }}
  name    = var.rName
  feature = "CUSTOM_LABELS"

{{- template "tags" . }}
}
