# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = var.rName

  firewall_policy {}
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
