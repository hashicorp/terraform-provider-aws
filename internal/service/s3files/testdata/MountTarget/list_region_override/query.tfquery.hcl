# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_s3files_mount_target" "test" {
  provider = aws

  config {
    region         = var.region
    file_system_id = aws_s3files_file_system.test.id
  }
}
