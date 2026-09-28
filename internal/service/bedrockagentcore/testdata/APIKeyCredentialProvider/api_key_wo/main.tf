# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_api_key_credential_provider" "test" {
  name               = var.rName
  api_key_wo         = var.api_key_wo
  api_key_wo_version = var.api_key_wo_version
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "api_key_wo" {
  type     = string
  nullable = false
}

variable "api_key_wo_version" {
  type     = number
  nullable = false
}