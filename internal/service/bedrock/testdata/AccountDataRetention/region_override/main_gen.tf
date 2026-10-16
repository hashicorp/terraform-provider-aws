# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrock_account_data_retention" "test" {
  region = var.region

  mode = "default"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
