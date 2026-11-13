# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sesv2_tenant" "test" {
  region = var.region

  tenant_name = var.rName
}

resource "aws_sesv2_tenant_suppression_attributes" "test" {
  region = var.region

  tenant_name        = aws_sesv2_tenant.test.tenant_name
  suppression_scope  = "TENANT"
  suppressed_reasons = ["BOUNCE", "COMPLAINT"]
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
