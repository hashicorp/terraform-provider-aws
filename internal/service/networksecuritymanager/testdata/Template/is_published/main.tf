# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_template" "test" {
  name          = var.rName
  firewall_type = "WAF"
  is_published  = var.is_published
  rule_arns     = slice(aws_networksecuritymanager_rule.test[*].arn, 0, var.rule_count)
}

resource "aws_networksecuritymanager_rule" "test" {
  count = 2

  name          = "${var.rName}-${count.index}"
  firewall_type = "WAF"
  type          = "CONFIGURATION"
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

variable "is_published" {
  type     = bool
  nullable = false
}

variable "rule_count" {
  type     = number
  nullable = false
}
