resource "aws_dms_data_provider" "test" {
  data_provider_name = var.rName
  engine             = "postgres"

  settings {
    postgres_settings {
      server_name   = "example.com"
      port          = 5432
      database_name = "testdb"
      ssl_mode      = "none"
    }
  }

{{- template "tags" . }}
}
