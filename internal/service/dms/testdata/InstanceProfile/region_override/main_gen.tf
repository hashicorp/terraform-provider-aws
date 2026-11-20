# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dms_instance_profile" "test" {
  region = var.region

}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
