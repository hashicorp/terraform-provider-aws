# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_s3_bucket_public_access_block" "test" {
  provider = aws
}

list "aws_s3_bucket_public_access_block" "include" {
  provider = aws

  include_resource = true
}
