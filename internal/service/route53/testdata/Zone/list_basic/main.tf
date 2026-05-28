# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_zone" "test" {
  count = var.resource_count

  name = "subdomain${count.index}.${var.zoneName}"
}

variable "zoneName" {
  description = "Root zone name for hosted zones"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
