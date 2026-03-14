# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_s3_object" "test" {
  provider = aws

  config {
    bucket = aws_s3_bucket.test.bucket
  }

  include_resource = true
}
