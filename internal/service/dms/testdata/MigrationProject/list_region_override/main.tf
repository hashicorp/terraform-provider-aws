# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dms_migration_project" "test" {
  count  = var.resource_count
  region = var.region

  name                 = "${var.rName}-${count.index}"
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

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_dms_instance_profile" "test" {
  region = var.region
}

data "aws_partition" "current" {}

data "aws_region" "current" {
  region = var.region
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
  region                  = var.region
  name                    = "${var.rName}-source"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "source" {
  region        = var.region
  secret_id     = aws_secretsmanager_secret.source.id
  secret_string = jsonencode({ username = "example", password = "Example123456" })
}

resource "aws_secretsmanager_secret" "target" {
  region                  = var.region
  name                    = "${var.rName}-target"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "target" {
  region        = var.region
  secret_id     = aws_secretsmanager_secret.target.id
  secret_string = jsonencode({ username = "example", password = "Example123456" })
}

resource "aws_dms_data_provider" "source" {
  region = var.region
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
  region = var.region
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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
