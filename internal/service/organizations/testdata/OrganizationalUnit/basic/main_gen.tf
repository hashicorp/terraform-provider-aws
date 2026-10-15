# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_organizations_organizational_unit" "test" {
  name      = var.rName
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

data "aws_organizations_organization" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
