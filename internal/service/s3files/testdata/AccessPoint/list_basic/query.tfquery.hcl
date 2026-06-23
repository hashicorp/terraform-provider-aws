# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_s3files_access_point" "test" {
  provider = aws

  config {
    file_system_id = aws_s3files_file_system.test.id
  }
}
