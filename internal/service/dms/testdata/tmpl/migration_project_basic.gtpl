resource "aws_dms_migration_project" "test" {
{{- template "region" }}
  instance_profile_arn = aws_dms_instance_profile.test.arn

  source_data_provider_descriptor {
    data_provider_arn = aws_dms_data_provider.source.arn
  }

  target_data_provider_descriptor {
    data_provider_arn = aws_dms_data_provider.target.arn
  }
{{- template "tags" . }}
}

resource "aws_dms_instance_profile" "test" {
{{- template "region" }}
}

resource "aws_dms_data_provider" "source" {
{{- template "region" }}
  engine = "postgres"

  settings {
    postgresql_settings {
      database_name = "example"
      port          = 5432
      server_name   = "${var.rName}-source.example.com"
      ssl_mode      = "none"
    }
  }
}

resource "aws_dms_data_provider" "target" {
{{- template "region" }}
  engine = "postgres"

  settings {
    postgresql_settings {
      database_name = "example"
      port          = 5432
      server_name   = "${var.rName}-target.example.com"
      ssl_mode      = "none"
    }
  }
}
