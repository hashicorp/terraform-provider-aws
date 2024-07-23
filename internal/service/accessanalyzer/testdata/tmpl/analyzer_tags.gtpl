resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = var.rName
{{- template "tags" . }}
}
