# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_efs_mount_target" "test" {
  provider = aws

  include_resource = true

  config {
    file_system_id = aws_efs_file_system.test.id
  }
}
