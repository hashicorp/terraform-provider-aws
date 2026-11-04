# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_template" "test" {
  name          = var.rName
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.draft.arn]
}

resource "aws_networksecuritymanager_rule" "draft" {
  name          = var.rName
  firewall_type = "WAF"
  type          = "CONFIGURATION"
  is_published  = false
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
