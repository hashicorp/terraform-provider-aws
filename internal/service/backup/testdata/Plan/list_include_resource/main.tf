# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0
#
resource "aws_backup_plan" "test" {
  count = var.resource_count
  name  = "${var.rName}-${count.index}"

  rule {
    rule_name         = var.rName
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }

  tags = var.resource_tags
}

resource "aws_backup_vault" "test" {
  name = var.rName
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

variable "resource_tags" {
  description = "Tags to set on resource"
  type        = map(string)
  nullable    = false
}
