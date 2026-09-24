# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_verifiedpermissions_policy_store_alias" "test" {
  provider = aws

  config {
    policy_store_id = aws_verifiedpermissions_policy_store.test.policy_store_id
  }
}