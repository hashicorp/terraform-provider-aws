# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_scope" "test" {
  name         = var.rName
  is_published = var.is_published

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

      include {
        expression {
          criteria {
            tags = {
              Name = var.tag_value
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

variable "is_published" {
  type     = bool
  nullable = false
}

variable "tag_value" {
  type     = string
  nullable = false
}
