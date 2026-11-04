# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_policy" "test" {
  name          = var.rName
  firewall_type = "SHIELD_ADVANCED"
  priority      = parseint(substr(sha1(var.rName), 0, 6), 16) + 1
  is_published  = var.is_published

  configuration {
    remediation_enabled = var.remediation_enabled
    resources_clean_up  = true
  }
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

variable "remediation_enabled" {
  type     = bool
  nullable = false
}
