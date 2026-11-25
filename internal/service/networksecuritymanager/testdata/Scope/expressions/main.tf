# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_scope" "test" {
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

      include {
        expression {
          and {
            tags = {
              Name = var.rName
            }
          }
          and {
            alb_config {
              ip_address_type = "ipv4"
              scheme          = "internet-facing"
            }
          }
        }
      }
    }

    resource_scope {
      resource_type = "AWS::EC2::EIP"

      include {
        expression {
          or {
            tags = {
              Name = var.rName
            }
          }
          or {
            tags = {}
          }
        }
      }
    }

    resource_scope {
      resource_type = "AWS::ApiGateway::Stage"

      exclude {
        expression {
          not {
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
