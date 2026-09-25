# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_template" "test" {
  name          = var.rName
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.test.arn]

  tags = var.resource_tags
}

resource "aws_networksecuritymanager_rule" "test" {
  name          = var.rName
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
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

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
