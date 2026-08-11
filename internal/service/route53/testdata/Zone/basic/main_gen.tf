# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_zone" "test" {
  comment = var.rName
  name    = "${var.zoneName}."
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
variable "zoneName" {
  type     = string
  nullable = false
}

