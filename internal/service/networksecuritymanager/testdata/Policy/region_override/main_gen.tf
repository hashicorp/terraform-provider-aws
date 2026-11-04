# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_policy" "test" {
  region = var.region

  name          = var.rName
  firewall_type = "SHIELD_ADVANCED"
  priority      = parseint(substr(sha1(var.rName), 0, 6), 16) + 1

  configuration {
    remediation_enabled = false
    resources_clean_up  = true
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
