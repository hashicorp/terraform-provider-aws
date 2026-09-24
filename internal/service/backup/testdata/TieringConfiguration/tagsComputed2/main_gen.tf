# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_backup_tiering_configuration" "test" {
  backup_vault_name          = aws_backup_vault.test.name
  tiering_configuration_name = var.rName

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 90
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

resource "aws_backup_vault" "test" {
  name = var.rName
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
