resource "aws_athena_system_table_configuration" "test" {
{{- template "region" }}
{{- template "tags" . }}
}
