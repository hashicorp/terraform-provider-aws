# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_resolver_rule" "test" {
  count  = var.resource_count
  region = var.region

  domain_name = "${count.index}.${var.domain}"
  name        = "${var.rName}-${count.index}"
  rule_type   = "SYSTEM"
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
