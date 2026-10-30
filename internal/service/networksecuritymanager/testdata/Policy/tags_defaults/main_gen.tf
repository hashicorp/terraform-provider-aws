# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

resource "aws_networksecuritymanager_policy" "test" {
  name          = var.rName
  firewall_type = "SHIELD_ADVANCED"
  priority      = parseint(substr(sha1(var.rName), 0, 6), 16) + 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = true
  }

  tags = var.resource_tags
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

variable "provider_tags" {
  type     = map(string)
  nullable = false
}
