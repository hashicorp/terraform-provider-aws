data "aws_efs_file_system" "test" {
{{- template "region" }}
  file_system_id = aws_efs_file_system.test.id
}