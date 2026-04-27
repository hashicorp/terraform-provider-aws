# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_managed_resource_visibility" "test" {
  region = var.region

  default_visibility = "hidden"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
