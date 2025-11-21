# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehub_app" "test" {
  name = var.rName

  app_template {
    version = "2.0"

    app_component {
      name           = "appcommon"
      type           = "AWS::ResilienceHub::AppCommonAppComponent"
      resource_names = []
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
