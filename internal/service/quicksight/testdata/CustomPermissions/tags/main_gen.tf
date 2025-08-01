# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_quicksight_custom_permissions" "test" {
  custom_permissions_name = var.rName

  capabilities {
    print_reports    = "DENY"
    share_dashboards = "DENY"
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
