# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_iam_role_policy" "test" {
  provider = aws
  config {
    role_name = aws_iam_role.test.name
  }

  include_resource = true
}
