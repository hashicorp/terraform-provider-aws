# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_efs_file_system" "test" {
  file_system_id = aws_efs_file_system.test.id
}

resource "aws_efs_file_system" "test" {}
