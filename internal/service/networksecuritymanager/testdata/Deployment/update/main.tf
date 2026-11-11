# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_deployment" "test" {
  name         = var.rName
  description  = var.description
  is_published = var.is_published
  policy_arns  = [for i in var.policy_indexes : aws_networksecuritymanager_policy.test[i].arn]
  scope_arn    = aws_networksecuritymanager_scope.test[var.scope_index].arn

  configuration {
    enable_cross_account_visibility = false
  }
}

resource "aws_networksecuritymanager_rule" "test" {
  name          = var.rName
  firewall_type = "WAF"
  type          = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_policy" "test" {
  count = 2

  name          = "${var.rName}-${count.index}"
  firewall_type = "WAF"
  priority      = parseint(substr(sha1("${var.rName}-${count.index}"), 0, 6), 16) + 1

  template_or_rule {
    rule_identifier = aws_networksecuritymanager_rule.test.arn
  }

  configuration {
    remediation_enabled = false
    resources_clean_up  = true

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }
}

resource "aws_networksecuritymanager_scope" "test" {
  count = 2

  name = "${var.rName}-${count.index}"

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"
      include_all   = true
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "description" {
  type    = string
  default = null
}

variable "is_published" {
  type    = bool
  default = null
}

variable "policy_indexes" {
  type     = list(number)
  nullable = false
}

variable "scope_index" {
  type     = number
  nullable = false
}
