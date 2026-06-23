# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_workmail_organization" "test" {
  organization_alias = var.rName
  delete_directory   = true
}

resource "aws_workmail_default_domain" "test" {
  organization_id = aws_workmail_organization.test.organization_id
  domain_name     = aws_workmail_organization.test.default_mail_domain
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
