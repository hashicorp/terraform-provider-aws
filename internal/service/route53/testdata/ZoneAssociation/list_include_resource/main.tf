# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_zone_association" "test" {
  count = var.resource_count

  zone_id = aws_route53_zone.foo[count.index].id
  vpc_id  = aws_vpc.bar.id
}

resource "aws_vpc" "bar" {
  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = var.rName
  }
}

resource "aws_vpc" "foo" {
  count = var.resource_count

  cidr_block           = cidrsubnet("10.0.0.0/8", 8, count.index)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.rName}-${count.index}"
  }
}

resource "aws_route53_zone" "foo" {
  count = var.resource_count

  name = "${var.rName}-${count.index}.example.com"

  vpc {
    vpc_id = aws_vpc.foo[count.index].id
  }

  lifecycle {
    ignore_changes = [vpc]
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
