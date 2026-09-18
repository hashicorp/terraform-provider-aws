# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_xray_resource_policy" "test" {
  region = var.region

  policy_name                 = var.rName
  bypass_policy_lockout_check = true

  policy_document = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowXRayAccess",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
        "xray:*",
        "xray:PutResourcePolicy"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
