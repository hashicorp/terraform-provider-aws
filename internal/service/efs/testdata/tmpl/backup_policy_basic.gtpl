resource "aws_efs_backup_policy" "test" {
{{- template "region" }}
  file_system_id = aws_efs_file_system.test.id

  backup_policy {
    status = "ENABLED"
  }
}

resource "aws_efs_file_system" "test" {
{{- template "region" }}
}
