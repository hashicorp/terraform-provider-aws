# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_wafv2_ip_set" "test" {
  name               = var.rName
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
