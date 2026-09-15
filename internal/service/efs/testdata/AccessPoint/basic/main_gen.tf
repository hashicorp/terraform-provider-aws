# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
}

resource "aws_efs_file_system" "test" {
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
