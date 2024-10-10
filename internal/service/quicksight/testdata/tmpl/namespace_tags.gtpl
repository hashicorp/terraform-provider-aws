resource "aws_quicksight_namespace" "test" {
  namespace = var.rName
{{- template "tags" . }}
}
