# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_efs_file_system" "test" {
  creation_token = aws_efs_file_system.test.creation_token
}

resource "aws_efs_file_system" "test" {
  creation_token = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}