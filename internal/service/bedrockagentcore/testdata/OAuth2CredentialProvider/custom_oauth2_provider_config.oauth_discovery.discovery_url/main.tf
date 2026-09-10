# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = var.rName

  credential_provider_vendor = "CustomOauth2"

  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id_wo                  = var.client_id
      client_secret_wo              = var.client_secret
      client_credentials_wo_version = var.client_credentials_version

      oauth_discovery {
        discovery_url = var.discovery_url
      }
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "client_id" {
  type     = string
  nullable = false
}

variable "client_secret" {
  type     = string
  nullable = false
}

variable "client_credentials_version" {
  type     = number
  nullable = false
}

variable "discovery_url" {
  type     = string
  nullable = false
}