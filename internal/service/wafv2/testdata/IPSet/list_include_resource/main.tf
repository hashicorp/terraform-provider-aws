# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_wafv2_ip_set" "test" {
  count = var.resource_count

  name               = "${var.rName}-${count.index}"
  description        = "${var.rName}-${count.index}"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32"]

  tags = {
    key1 = "value1"
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
