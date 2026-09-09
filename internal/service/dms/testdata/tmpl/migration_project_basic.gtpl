resource "aws_dms_migration_project" "test" {
{{- template "region" }}
  instance_profile_arn = aws_dms_instance_profile.test.arn

  source_data_provider_descriptor {
    data_provider_arn               = aws_dms_data_provider.source.arn
    secrets_manager_access_role_arn = aws_iam_role.test.arn
    secrets_manager_secret_id       = aws_secretsmanager_secret.source.arn
  }

  target_data_provider_descriptor {
    data_provider_arn               = aws_dms_data_provider.target.arn
    secrets_manager_access_role_arn = aws_iam_role.test.arn
    secrets_manager_secret_id       = aws_secretsmanager_secret.target.arn
  }
{{- template "tags" . }}

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_dms_instance_profile" "test" {
{{- template "region" }}
}

data "aws_partition" "current" {}

data "aws_region" "current" {
{{- template "region" -}}
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "dms.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "secretsmanager:*"
      Effect   = "Allow"
      Resource = "*"
    }]
  })
}

resource "aws_secretsmanager_secret" "source" {
{{- template "region" }}
  name                    = "${var.rName}-source"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "source" {
{{- template "region" }}
  secret_id     = aws_secretsmanager_secret.source.id
  secret_string = jsonencode({ username = "example", password = "Example123456" })
}

resource "aws_secretsmanager_secret" "target" {
{{- template "region" }}
  name                    = "${var.rName}-target"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "target" {
{{- template "region" }}
  secret_id     = aws_secretsmanager_secret.target.id
  secret_string = jsonencode({ username = "example", password = "Example123456" })
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
