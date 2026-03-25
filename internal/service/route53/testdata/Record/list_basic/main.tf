# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_zone" "test" {
  name = "${var.rName}.com"
}

resource "aws_route53_record" "test" {
  count = var.resource_count

  zone_id = aws_route53_zone.test.zone_id
  name    = "${var.rName}-${count.index}.${aws_route53_zone.test.name}"
  type    = "A"
  ttl     = 300
  records = ["10.0.0.${count.index}"]
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
