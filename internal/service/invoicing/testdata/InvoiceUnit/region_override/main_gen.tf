# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_invoicing_invoice_unit" "test" {
  region = var.region

  name             = var.rName
  invoice_receiver = data.aws_caller_identity.current.account_id

  rule {
    linked_accounts = [data.aws_caller_identity.current.account_id]
  }
}

data "aws_caller_identity" "current" {}

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
