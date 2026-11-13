# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_member" "test" {
  depends_on = [aws_securityhub_account.test]
  account_id = "111111111111"
}

resource "aws_securityhub_account" "test" {
}

