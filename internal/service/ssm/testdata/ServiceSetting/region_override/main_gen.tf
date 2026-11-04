# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_service_setting" "test" {
  region = var.region

  setting_id    = "/ssm/parameter-store/high-throughput-enabled"
  setting_value = "false"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
