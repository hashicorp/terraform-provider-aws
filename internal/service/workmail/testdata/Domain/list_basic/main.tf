# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_workmail_organization" "test" {
  organization_alias = var.rName
  delete_directory   = true
}

resource "aws_workmail_domain" "test" {
  count = var.resource_count

  organization_id = aws_workmail_organization.test.organization_id
  domain_name     = "${var.rName}-${count.index}.example.com"
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
