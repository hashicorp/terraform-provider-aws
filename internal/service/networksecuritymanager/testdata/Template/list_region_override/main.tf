# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_template" "test" {
  count  = var.resource_count
  region = var.region

  name          = "${var.rName}-${count.index}"
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.test.arn]
}

resource "aws_networksecuritymanager_rule" "test" {
  region = var.region

  name          = var.rName
  firewall_type = "WAF"
  type          = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
