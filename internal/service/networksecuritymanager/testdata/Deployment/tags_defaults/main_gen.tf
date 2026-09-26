# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

resource "aws_networksecuritymanager_deployment" "test" {
  name        = var.rName
  policy_arns = [aws_networksecuritymanager_policy.test.arn]
  scope_arn   = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }

  tags = var.resource_tags
}

resource "aws_networksecuritymanager_policy" "test" {
  name          = var.rName
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.test.arn
  }
}

resource "aws_networksecuritymanager_rule" "test" {
  name          = var.rName
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_scope" "test" {
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"

      include {
        expression {
          criteria {
            tags = {
              nsm-acctest = var.rName
            }
          }
        }
      }
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

variable "provider_tags" {
  type     = map(string)
  nullable = false
}
