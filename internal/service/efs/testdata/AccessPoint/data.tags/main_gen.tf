# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_efs_access_point" "test" {
  access_point_id = aws_efs_access_point.test.id
}
resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  tags = var.resource_tags
}

resource "aws_efs_file_system" "test" {
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
