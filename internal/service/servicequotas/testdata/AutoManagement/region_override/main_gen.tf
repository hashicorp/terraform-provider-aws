# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_servicequotas_auto_management" "test" {
  region = var.region


  opt_in_level = "ACCOUNT"
  opt_in_type  = "NotifyOnly"
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
