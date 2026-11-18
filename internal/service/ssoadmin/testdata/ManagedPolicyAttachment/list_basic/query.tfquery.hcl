# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_ssoadmin_managed_policy_attachment" "test" {
  provider = aws

  config {
    instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
    permission_set_arn = aws_ssoadmin_permission_set.test.arn
  }
}
