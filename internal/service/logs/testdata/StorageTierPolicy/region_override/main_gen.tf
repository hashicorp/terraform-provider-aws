# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_storage_tier_policy" "test" {
  region = var.region

  storage_tier = "INTELLIGENT_TIERING"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
