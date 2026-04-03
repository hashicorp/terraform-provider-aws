# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networkfirewall_proxy_configuration" "test" {
  region = var.region
  count  = var.resource_count

  name = "${var.rName}-${count.index}"

  default_rule_phase_actions {
    post_response = "ALLOW"
    pre_dns       = "ALLOW"
    pre_request   = "ALLOW"
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
