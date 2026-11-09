# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_quicksight_custom_permissions" "test" {
  custom_permissions_name = var.rName

  capabilities {
    print_reports    = "DENY"
    share_dashboards = "DENY"
  }
}

resource "aws_quicksight_account_custom_permission" "test" {
  custom_permissions_name = aws_quicksight_custom_permissions.test.custom_permissions_name
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
