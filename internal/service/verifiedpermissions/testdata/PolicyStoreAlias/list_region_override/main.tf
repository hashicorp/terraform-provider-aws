# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_verifiedpermissions_policy_store" "test" {
  region = var.region

  description = var.rName

  validation_settings {
    mode = "OFF"
  }
}

resource "aws_verifiedpermissions_policy_store_alias" "test" {
  count  = var.resource_count
  region = var.region

  alias_name      = "policy-store-alias/${var.rName}-${count.index}"
  policy_store_id = aws_verifiedpermissions_policy_store.test.policy_store_id
}

variable "rName" {
  description = "Name for the test resources"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of aliases to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region in which to create the resources"
  type        = string
  nullable    = false
}