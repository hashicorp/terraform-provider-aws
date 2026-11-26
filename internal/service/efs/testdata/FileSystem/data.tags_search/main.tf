# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_efs_file_system" "test" {
  tags = aws_efs_file_system.test.tags

  depends_on = [aws_efs_file_system.wrong-env, aws_efs_file_system.no-tags]
}

resource "aws_efs_file_system" "test" {
  tags = {
    Name        = "default-efs"
    Environment = "dev"
  }
}

resource "aws_efs_file_system" "wrong-env" {
  tags = {
    Environment = "test"
  }
}

resource "aws_efs_file_system" "no-tags" {}
