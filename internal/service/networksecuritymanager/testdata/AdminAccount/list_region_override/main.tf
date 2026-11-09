# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1
  region     = var.region
}

data "aws_caller_identity" "current" {}

variable "region" {
  description = "Region to manage the admin account in"
  type        = string
  nullable    = false
}
