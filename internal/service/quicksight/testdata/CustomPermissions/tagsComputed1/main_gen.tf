# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_quicksight_custom_permissions" "test" {
  custom_permissions_name = var.rName

  capabilities {
    print_reports    = "DENY"
    share_dashboards = "DENY"
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
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
