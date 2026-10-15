# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = "*.${var.zoneName}"
  type    = "A"
  ttl     = 300
  records = ["10.0.0.0"]
}

resource "aws_route53_zone" "test" {
  name = var.zoneName
}

variable "zoneName" {
  type     = string
  nullable = false
}
