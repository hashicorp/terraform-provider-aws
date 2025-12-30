# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_resolver_rule_association" "test" {
  region = var.region

  name             = var.rName
  resolver_rule_id = aws_route53_resolver_rule.test.id
  vpc_id           = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_route53_resolver_rule" "test" {
  region = var.region

  domain_name = var.domain
  name        = var.rName
  rule_type   = "SYSTEM"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
variable "domain" {
  type     = string
  nullable = false
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
