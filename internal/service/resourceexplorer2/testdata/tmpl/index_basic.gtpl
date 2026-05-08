resource "aws_resourceexplorer2_index" "test" {
{{- template "region" }}
  type = "LOCAL"
{{- template "tags" . }}
}
