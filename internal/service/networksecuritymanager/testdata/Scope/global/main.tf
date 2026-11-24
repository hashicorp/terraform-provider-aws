# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networksecuritymanager_scope" "test" {
  region = var.region

  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"
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
