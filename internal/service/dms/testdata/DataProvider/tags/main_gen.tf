# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dms_data_provider" "test" {

  engine = "postgres"

  settings {
    postgresql_settings {
      database_name = "example"
      port          = 5432
      server_name   = "example.com"
      ssl_mode      = "none"
    }
  }

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
