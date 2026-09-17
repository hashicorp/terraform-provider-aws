# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_workmail_organization" "test" {
  organization_alias = var.rName
  delete_directory   = true
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
