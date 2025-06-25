# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ebs_snapshot_block_public_access" "test" {
  region = var.region

  state = "block-all-sharing"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
