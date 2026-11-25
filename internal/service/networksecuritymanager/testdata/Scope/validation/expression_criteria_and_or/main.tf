# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Invalid: expression sets both criteria and or.
resource "aws_networksecuritymanager_scope" "test" {
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

      include {
        expression {
          criteria {
            tags = {
              Name = var.rName
            }
          }

          or {
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
