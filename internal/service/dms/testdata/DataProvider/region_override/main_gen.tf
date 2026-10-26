# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dms_data_provider" "test" {
  region = var.region


  engine = "postgres"

  settings {
    postgresql_settings {
      database_name = "example"
      port          = 5432
      server_name   = "${var.rName}.example.com"
      ssl_mode      = "none"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
