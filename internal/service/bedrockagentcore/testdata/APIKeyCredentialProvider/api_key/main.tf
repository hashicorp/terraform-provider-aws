# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_api_key_credential_provider" "test" {
  name    = var.rName
  api_key = var.api_key
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "api_key" {
  type     = string
  nullable = false
}