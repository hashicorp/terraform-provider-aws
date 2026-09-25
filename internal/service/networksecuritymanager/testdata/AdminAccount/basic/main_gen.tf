# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_caller_identity" "current" {}

resource "aws_organizations_aws_service_access" "test" {
  service_principal = "network-security-manager.amazonaws.com"
}

resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = 1
}

