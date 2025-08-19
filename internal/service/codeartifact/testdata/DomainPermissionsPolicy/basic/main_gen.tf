# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_codeartifact_domain_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "codeartifact:CreateRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}

resource "aws_codeartifact_domain" "test" {
  domain         = var.rName
  encryption_key = aws_kms_key.test.arn
}

resource "aws_kms_key" "test" {
  description             = var.rName
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
