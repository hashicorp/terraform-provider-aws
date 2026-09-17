# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_resolver_rule_association" "test" {
  count = var.resource_count

  region           = var.region
  name             = "${var.rName}-${count.index}"
  resolver_rule_id = aws_route53_resolver_rule.test[count.index].id
  vpc_id           = aws_vpc.test.id
}

resource "aws_route53_resolver_rule" "test" {
  count = var.resource_count

  region      = var.region
  domain_name = "${count.index}.${var.domain}"
  name        = "${var.rName}-rule-${count.index}"
  rule_type   = "SYSTEM"
}

resource "aws_vpc" "test" {
  region               = var.region
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
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

variable "domain" {
  description = "Domain name for resolver rule"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
