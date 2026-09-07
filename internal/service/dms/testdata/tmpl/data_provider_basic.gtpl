resource "aws_dms_data_provider" "test" {
{{- template "region" }}

  name   = var.rName
  engine = "postgres"

  settings {
    postgresql_settings {
      database_name = "example"
      port          = 5432
      server_name   = "example.com"
      ssl_mode      = "none"
    }
  }

{{- template "tags" . }}
}
