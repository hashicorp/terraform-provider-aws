# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_auditmanager_account_registration" "test" {
  region = var.region
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
