# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_workmail_organization" "test" {
  region = var.region

  organization_alias = var.rName
  delete_directory   = true
}

resource "aws_workmail_domain" "test" {
  region = var.region

  organization_id = aws_workmail_organization.test.organization_id
  domain_name     = "${var.rName}.example.com"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
