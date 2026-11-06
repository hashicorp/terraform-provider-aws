# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_deployment" "test" {
  region = var.region

  name        = var.rName
  policy_arns = [aws_networksecuritymanager_policy.test.arn]
  scope_arn   = aws_networksecuritymanager_scope.test.arn

  configuration {
    enable_cross_account_visibility = false
  }
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

resource "aws_networksecuritymanager_policy" "test" {
  region = var.region

  name          = var.rName
  firewall_type = "WAF"
  priority      = parseint(substr(sha1(var.rName), 0, 6), 16) + 1

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
  region = var.region

  name = var.rName

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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
