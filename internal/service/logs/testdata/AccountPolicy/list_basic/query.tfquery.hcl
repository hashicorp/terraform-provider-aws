# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_cloudwatch_log_account_policy" "test" {
  provider = aws

  config {
    policy_type = "RETENTION_POLICY"
  }
}
