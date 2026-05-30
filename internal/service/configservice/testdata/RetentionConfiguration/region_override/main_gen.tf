# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_config_retention_configuration" "test" {
  region = var.region

  retention_period_in_days = 90
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
