resource "aws_efs_access_point" "test" {
{{- template "region" }}
  file_system_id = aws_efs_file_system.test.id

{{- template "tags" . }}
}

resource "aws_efs_file_system" "test" {
{{- template "region" }}
}
