# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_rds_certificate" "test" {
  region = var.region

  certificate_identifier = "rds-ca-rsa4096-g1"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
