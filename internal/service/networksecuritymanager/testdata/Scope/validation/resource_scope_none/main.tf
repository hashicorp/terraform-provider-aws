# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Invalid: resource_scope sets none of include_all, include and exclude.
resource "aws_networksecuritymanager_scope" "test" {
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
