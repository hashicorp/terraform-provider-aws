# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_devopsguru_event_sources_config" "test" {
  region = var.region

  event_sources {
    amazon_code_guru_profiler {
      status = "ENABLED"
    }
  }
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
