# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_appfabric_app_bundle" "test" {
  region = var.region

}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
