# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1
}

data "aws_caller_identity" "current" {}

