# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_template" "test" {
  name          = coalesce(var.template_name, var.rName)
  description   = var.description
  firewall_type = "WAF"
  is_published  = var.is_published
  rule_arns     = [for k in var.rules : local.rule_arns[k]]
}

resource "aws_networksecuritymanager_rule" "a" {
  name          = "${var.rName}-a"
  firewall_type = "WAF"
  type          = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_rule" "b" {
  name          = "${var.rName}-b"
  firewall_type = "WAF"
  type          = "CONFIGURATION"
  configuration = jsonencode({
    VisibilityConfig = {
      SampledRequestsEnabled   = true
      CloudWatchMetricsEnabled = true
      MetricName               = "tf-acc-test"
    }
  })
}

resource "aws_networksecuritymanager_rule" "c" {
  name          = "${var.rName}-c"
  firewall_type = "WAF"
  type          = "INSPECTION"
  configuration = jsonencode({
    PreProcessFirewallManagerRuleGroups = [{
      Name = "AWSManagedRulesCommonRuleSet"
      FirewallManagerStatement = {
        ManagedRuleGroupStatement = {
          VendorName = "AWS"
          Name       = "AWSManagedRulesCommonRuleSet"
        }
      }
      OverrideAction = {
        None = {}
      }
      VisibilityConfig = {
        SampledRequestsEnabled   = true
        CloudWatchMetricsEnabled = true
        MetricName               = "AWSManagedRulesCommonRuleSet"
      }
    }]
  })
}

locals {
  rule_arns = {
    a = aws_networksecuritymanager_rule.a.arn
    b = aws_networksecuritymanager_rule.b.arn
    c = aws_networksecuritymanager_rule.c.arn
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "template_name" {
  description = "Template name, when it differs from rName"
  type        = string
  default     = null
}

variable "description" {
  type    = string
  default = null
}

variable "is_published" {
  type    = bool
  default = null
}

variable "rules" {
  description = "Keys of the rules the template references, in order"
  type        = list(string)
  nullable    = false
}
