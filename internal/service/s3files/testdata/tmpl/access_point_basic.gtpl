resource "aws_s3files_access_point" "test" {
  file_system_id = aws_s3files_file_system.test.file_system_id

{{- template "tags" . }}
}
