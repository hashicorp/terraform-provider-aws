# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_scope" "test" {
  name                   = var.rName
  specialized_scope_type = "IGW_FIREWALL_SUBNET_CONNECTIVITY_OVERRIDE"

  scope_configuration {
    resource_scope {
      resource_type = "AWS::EC2::Subnet"

      include {
        expression {
          criteria {
            tags = {
              Name = var.rName
            }
          }
        }
      }
    }
  }
}

resource "aws_networksecuritymanager_rule" "test" {
  name          = var.rName
  firewall_type = "IGW_FIREWALL"
  type          = "INSPECTION"
  configuration = jsonencode({
    FirewallDefaultAction = "ACCEPT"
  })
}

resource "aws_networksecuritymanager_policy" "test" {
  name          = var.rName
  firewall_type = "IGW_FIREWALL"
  priority      = parseint(substr(sha1(var.rName), 0, 6), 16) + 1

  template_or_rule {
    rule_identifier = aws_networksecuritymanager_rule.test.arn
  }

  configuration {
    # The service requires remediation_enabled and resources_clean_up to both be true for IGW_FIREWALL policies.
    remediation_enabled = true
    resources_clean_up  = true

    igw_firewall_config {
      default_subnet_connectivity = var.default_subnet_connectivity

      dynamic "subnet_connectivity_override" {
        for_each = var.override_connectivity_type == null ? [] : [var.override_connectivity_type]
        content {
          connectivity_type = subnet_connectivity_override.value
          scope_arn         = [aws_networksecuritymanager_scope.test.arn]
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

variable "default_subnet_connectivity" {
  type     = string
  nullable = false
}

variable "override_connectivity_type" {
  type     = string
  nullable = true
  default  = null
}
