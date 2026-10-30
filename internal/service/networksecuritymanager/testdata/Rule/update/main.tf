# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_rule" "test" {
  name          = var.rName
  description   = var.description
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = var.configuration
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "description" {
  type     = string
  nullable = false
}

variable "configuration" {
  type     = string
  nullable = false
}
