resource "aws_detective_graph" "test" {
{{- template "region" }}
{{- template "tags" . }}
}