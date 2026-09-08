# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

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

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
