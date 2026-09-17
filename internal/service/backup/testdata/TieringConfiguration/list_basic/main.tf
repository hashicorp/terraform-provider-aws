# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_backup_tiering_configuration" "test" {
  count = var.resource_count

  name              = "${replace(var.rName, "-", "_")}_${count.index}"
  backup_vault_name = aws_backup_vault.test[count.index].name

  resource_selection {
    resource_type                 = "S3"
    resources                     = ["*"]
    tiering_down_settings_in_days = 60
  }
}

resource "aws_backup_vault" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
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
