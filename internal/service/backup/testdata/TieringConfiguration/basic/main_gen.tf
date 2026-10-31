# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_backup_tiering_configuration" "test" {
  name              = replace(var.rName, "-", "_")
  backup_vault_name = aws_backup_vault.test.name

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 60
  }
}

resource "aws_backup_vault" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
