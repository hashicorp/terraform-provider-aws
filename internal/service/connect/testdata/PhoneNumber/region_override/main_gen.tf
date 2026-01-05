# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_connect_phone_number" "test" {
  region = var.region

  target_arn   = aws_connect_instance.test.arn
  country_code = "US"
  type         = "DID"
}

resource "aws_connect_instance" "test" {
  region = var.region

  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = var.rName
  outbound_calls_enabled   = true
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
