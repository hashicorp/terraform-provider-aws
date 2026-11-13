# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_ssoadmin_account_assignment" "test" {
  provider = aws

  include_resource = true

  config {
    account_id         = data.aws_caller_identity.current.account_id
    instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
    permission_set_arn = aws_ssoadmin_permission_set.test.arn
  }
}
