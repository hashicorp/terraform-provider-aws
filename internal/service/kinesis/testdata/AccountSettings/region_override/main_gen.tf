# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_kinesis_account_settings" "test" {
  region = var.region

  minimum_throughput_billing_commitment {
    status = "DISABLED"
  }
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
