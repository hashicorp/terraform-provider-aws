# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = var.rName
  parent_id = data.aws_organizations_organization.current.roots[0].id

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

data "aws_organizations_organization" "current" {}

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
