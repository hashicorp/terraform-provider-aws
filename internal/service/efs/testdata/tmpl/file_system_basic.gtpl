resource "aws_efs_file_system" "test" {
{{- template "region" }}
{{- template "tags" . }}
}
