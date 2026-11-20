# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Invalid: account_filter sets both include_all and include.
resource "aws_networksecuritymanager_scope" "test" {
  name = var.rName

  scope_configuration {
    account_filter {
      include_all = true

      include {
        account_ids = ["123456789012"]
      }
    }

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
