# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networkfirewall_firewall_policy_v2" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"

  firewall_policy {}
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
