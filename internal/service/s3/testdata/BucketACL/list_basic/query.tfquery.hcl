# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_s3_bucket_acl" "test" {
  provider = aws
}

list "aws_s3_bucket_acl" "include" {
  provider = aws

  include_resource = true
}
