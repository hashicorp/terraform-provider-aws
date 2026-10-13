resource "aws_dsql_cluster" "test" {
{{- template "region" }}
{{- template "tags" . }}
}
