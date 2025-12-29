# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_image_block_public_access" "test" {
  region = var.region

  state = "block-new-sharing"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
