# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_record_set_protection" "test" {
  zone_id     = aws_route53_zone.test.zone_id
  record_name = var.recordName
  record_type = "A"
}

resource "aws_route53_zone" "test" {
  name = var.zoneName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
variable "recordName" {
  type     = string
  nullable = false
}

variable "zoneName" {
  type     = string
  nullable = false
}

