# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = var.rName

  credential_provider_vendor = "CustomOauth2"

  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id_wo                  = "token-exchange-client-id"
      client_secret_wo              = "token-exchange-client-secret"
      client_credentials_wo_version = 1
      client_authentication_method  = "CLIENT_SECRET_BASIC"

      oauth_discovery {
        discovery_url = "https://dev-example.auth0.com/.well-known/openid-configuration" # nosemgrep:ci.semgrep.domain-names.domain-names-tf
      }

      on_behalf_of_token_exchange_config {
        grant_type = "TOKEN_EXCHANGE"

        token_exchange_grant_type_config {
          actor_token_content = "M2M"
          actor_token_scopes  = ["read", "write"]
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
