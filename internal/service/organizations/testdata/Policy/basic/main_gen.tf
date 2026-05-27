# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_organizations_policy" "test" {
  name = var.rName

  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
