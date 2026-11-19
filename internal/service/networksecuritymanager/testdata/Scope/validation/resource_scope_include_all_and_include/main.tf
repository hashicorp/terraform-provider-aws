# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Invalid: resource_scope sets both include_all and include.
resource "aws_networksecuritymanager_scope" "test" {
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"
      include_all   = true

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
