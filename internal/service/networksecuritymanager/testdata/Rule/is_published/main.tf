# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_rule" "test" {
  name          = var.rName
  firewall_type = "WAF"
  type          = "CONFIGURATION"
  is_published  = var.is_published
  configuration = var.configuration
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

variable "configuration" {
  type     = string
  nullable = false
}
