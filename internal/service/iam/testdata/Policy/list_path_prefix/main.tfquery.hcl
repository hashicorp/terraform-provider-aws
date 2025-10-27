# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

list "aws_iam_policy" "expected" {
  provider = aws

  config {
    path_prefix = var.expected_path_name
  }
}
