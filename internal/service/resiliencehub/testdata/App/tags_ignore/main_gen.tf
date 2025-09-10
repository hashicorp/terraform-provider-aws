# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
  ignore_tags {
    keys = var.ignore_tag_keys
  }
}

resource "aws_resiliencehub_app" "test" {
  name                    = var.rName
  app_assessment_schedule = "Disabled"

  app_template {
    version = "2.0"

    resource {
      name = "test-resource"
      type = "AWS::Lambda::Function"
      logical_resource_id {
        identifier = "TestResource"
      }
    }

    app_component {
      name           = "appcommon"
      type           = "AWS::ResilienceHub::AppCommonAppComponent"
      resource_names = []
    }

    app_component {
      name           = "test-component"
      type           = "AWS::ResilienceHub::ComputeAppComponent"
      resource_names = ["test-resource"]
    }
  }

  tags = var.resource_tags
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
  nullable = true
  default  = null
}

variable "ignore_tag_keys" {
  type     = set(string)
  nullable = false
}
