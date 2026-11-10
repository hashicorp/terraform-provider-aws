# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_template" "test" {
  name          = var.rName
  description   = var.description
  firewall_type = "WAF"
  rule_arns     = slice(var.reversed ? reverse(aws_networksecuritymanager_rule.test[*].arn) : aws_networksecuritymanager_rule.test[*].arn, 0, var.rule_count)
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

variable "description" {
  type    = string
  default = null
}

variable "reversed" {
  type     = bool
  default  = false
  nullable = false
}

variable "rule_count" {
  type     = number
  nullable = false
}
