# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_workmail_organization" "test" {
  organization_alias = var.rName
  delete_directory   = true
}

resource "aws_workmail_user" "test" {
  organization_id = aws_workmail_organization.test.organization_id
  name            = var.rName
  display_name    = var.rName
  password        = "TestTest1234!"
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
