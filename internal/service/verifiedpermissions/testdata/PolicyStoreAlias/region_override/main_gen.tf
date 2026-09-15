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
  region = var.region

  alias_name      = "policy-store-alias/${var.rName}"
  policy_store_id = aws_verifiedpermissions_policy_store.test.policy_store_id
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
