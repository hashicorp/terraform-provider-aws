# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = var.rName

  credential_provider_vendor = "CustomOauth2"

  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id     = "test-client-id"
      client_secret = "test-client-secret"

      oauth_discovery {
        authorization_server_metadata {
          authorization_endpoint      = "https://example.com/authorize" # nosemgrep:ci.semgrep.domain-names.domain-names-tf
          issuer                      = "https://example.com"           # nosemgrep:ci.semgrep.domain-names.domain-names-tf
          response_types              = ["code"]
          token_endpoint              = "https://example.com/token" # nosemgrep:ci.semgrep.domain-names.domain-names-tf
          token_endpoint_auth_methods = var.token_endpoint_auth_methods
        }
      }
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "token_endpoint_auth_methods" {
  type     = list(string)
  nullable = false
}