# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dms_migration_project" "test" {
  instance_profile_arn = aws_dms_instance_profile.test.arn

  source_data_provider_descriptor {
    data_provider_arn = aws_dms_data_provider.source.arn
  }

  target_data_provider_descriptor {
    data_provider_arn = aws_dms_data_provider.target.arn
  }
}

resource "aws_dms_instance_profile" "test" {
}

resource "aws_dms_data_provider" "source" {
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
