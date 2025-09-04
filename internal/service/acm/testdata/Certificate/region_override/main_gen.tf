# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_acm_certificate" "test" {
  region = var.region

  certificate_body = var.certificate_pem
  private_key      = var.private_key_pem
}

variable "certificate_pem" {
  type     = string
  nullable = false
}

variable "private_key_pem" {
  type     = string
  nullable = false
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
