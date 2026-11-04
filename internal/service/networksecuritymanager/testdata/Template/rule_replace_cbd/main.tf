# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_template" "test" {
  name          = var.rName
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.a.arn]
}

resource "aws_networksecuritymanager_rule" "a" {
  name          = var.rule_name
  firewall_type = "WAF"
  type          = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })

  lifecycle {
    create_before_destroy = true
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "rule_name" {
  type     = string
  nullable = false
}
