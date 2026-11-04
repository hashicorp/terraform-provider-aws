# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Invalid: two resource_scope blocks select the same resource_type differently.
resource "aws_networksecuritymanager_scope" "test" {
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"
      include_all   = true
    }

    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
