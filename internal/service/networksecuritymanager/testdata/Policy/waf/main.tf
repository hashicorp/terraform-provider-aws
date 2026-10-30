# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_policy" "test" {
  name          = var.rName
  description   = var.description
  firewall_type = "WAF"
  priority      = var.priority

  dynamic "template_or_rule" {
    for_each = var.rule_arns
    content {
      rule_arn = template_or_rule.value
    }
  }

  policy_configuration {
    remediation_enabled = var.remediation_enabled
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = var.existing_customer_web_acl_resolution
    }
  }
}

variable "rName" {
  type     = string
  nullable = false
}

variable "description" {
  type     = string
  nullable = false
}

variable "priority" {
  type     = number
  nullable = false
}

variable "rule_arns" {
  type     = list(string)
  nullable = false
}

variable "remediation_enabled" {
  type     = bool
  nullable = false
}

variable "existing_customer_web_acl_resolution" {
  type     = string
  nullable = false
}
