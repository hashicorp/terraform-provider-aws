# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_policy" "test" {
  count  = var.resource_count
  region = var.region

  name          = "${var.rName}-${count.index}"
  firewall_type = "SHIELD_ADVANCED"
  priority      = parseint(substr(sha1("${var.rName}-${count.index}"), 0, 6), 16) + 1

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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
