# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_accountaccess_entitlement" "test" {
  provider = aws

  config {
    account_id      = split(":", aws_iam_role.test.arn)[4]
    application_arn = aws_accountaccess_application.test.arn
  }
}
