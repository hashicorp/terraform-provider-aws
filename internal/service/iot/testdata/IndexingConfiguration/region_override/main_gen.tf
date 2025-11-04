# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_iot_indexing_configuration" "test" {
  region = var.region

  thing_group_indexing_configuration {
    thing_group_indexing_mode = "OFF"
  }

  thing_indexing_configuration {
    thing_indexing_mode = "OFF"
  }
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
