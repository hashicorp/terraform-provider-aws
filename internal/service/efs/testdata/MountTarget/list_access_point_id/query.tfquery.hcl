# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_efs_mount_target" "test" {
  provider = aws

  config {
    access_point_id = aws_efs_access_point.test.id
  }
}
