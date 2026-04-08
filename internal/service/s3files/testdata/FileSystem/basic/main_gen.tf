# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3files_file_system" "test" {
  bucket   = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
