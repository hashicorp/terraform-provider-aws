# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatchomni_access_grant" "test" {
  name       = var.rName
  domain_id  = "d-0123456789"
  space_id   = "11111111-2222-4333-8444-555555555555"
  permission = "CUSTOM"

  principal {
    principal_type = "IAM_ROLE"
    principal_id   = "arn:aws:iam::123456789012:role/example"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
