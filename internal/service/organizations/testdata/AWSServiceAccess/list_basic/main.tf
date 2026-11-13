# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_organizations_aws_service_access" "test" {
  count = var.resource_count

  service_principal = local.service_principals[count.index]
}

locals {
  service_principals = ["tagpolicies.tag.amazonaws.com", "config.amazonaws.com", "ds.amazonaws.com"]
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
