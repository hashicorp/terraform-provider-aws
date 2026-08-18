# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_config_connector" "test" {
  region = var.region

  azure {
    client_identifier = "00000000-0000-0000-0000-000000000000"
    tenant_identifier = "11111111-1111-1111-1111-111111111111"
  }
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
