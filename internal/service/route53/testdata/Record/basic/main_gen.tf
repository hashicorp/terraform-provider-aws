# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = var.zoneName
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "127.0.0.27"]
}

resource "aws_route53_zone" "test" {
  name = var.recordName
}

variable "recordName" {
  type     = string
  nullable = false
}

variable "zoneName" {
  type     = string
  nullable = false
}

