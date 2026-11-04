# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_template" "test" {
  name          = var.rName
  description   = var.description
  firewall_type = var.firewall_type
  rule_arns     = var.no_rules ? [] : var.rule_arns
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

variable "firewall_type" {
  type    = string
  default = "WAF"
}

variable "rule_arns" {
  type     = list(string)
  nullable = false
}

variable "no_rules" {
  description = "Set rule_arns to an empty list"
  type        = bool
  default     = false
}
