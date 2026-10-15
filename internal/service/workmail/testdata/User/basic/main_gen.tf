# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_workmail_user" "test" {
  organization_id = aws_workmail_organization.test.organization_id
  email           = "${var.rName}@${aws_workmail_organization.test.default_mail_domain}"
  name            = var.rName
  display_name    = var.rName
  city            = "bangalore"
  office          = "hashicorp"
}

resource "aws_workmail_organization" "test" {
  organization_alias = var.rName
  delete_directory   = true
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
