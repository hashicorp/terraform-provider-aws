# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

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

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
