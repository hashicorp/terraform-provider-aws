# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_resolver_rule" "test" {
  domain_name = var.rName
  rule_type   = "SYSTEM"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
