resource "aws_efs_replication_configuration" "test" {
{{- template "region" }}
  source_file_system_id = aws_efs_file_system.test.id

  destination {
    region = var.destination_region
  }
}

resource "aws_efs_file_system" "test" {
{{- template "region" }}
}
