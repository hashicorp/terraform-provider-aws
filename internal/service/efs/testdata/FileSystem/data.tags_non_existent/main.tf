# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_efs_file_system" "test" {
  tags = {
    Name = "Does_Not_Exist"
  }
}
