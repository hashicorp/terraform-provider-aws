# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_acm_certificate" "test" {
  count  = var.resource_count
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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
