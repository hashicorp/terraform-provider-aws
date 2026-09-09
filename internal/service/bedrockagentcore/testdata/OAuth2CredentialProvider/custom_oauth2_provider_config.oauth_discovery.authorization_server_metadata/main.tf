# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = var.rName

  credential_provider_vendor = "CustomOauth2"

  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id_wo                  = "keycloak-client-id"
      client_secret_wo              = "keycloak-client-secret"
      client_credentials_wo_version = 1

      oauth_discovery {
        authorization_server_metadata {
          issuer                 = "https://auth.company.com/realms/production"                               # nosemgrep:ci.semgrep.domain-names.domain-names-tf
          authorization_endpoint = "https://auth.company.com/realms/production/protocol/openid-connect/auth"  # nosemgrep:ci.semgrep.domain-names.domain-names-tf
          token_endpoint         = "https://auth.company.com/realms/production/protocol/openid-connect/token" # nosemgrep:ci.semgrep.domain-names.domain-names-tf
          response_types         = ["code", "id_token"]
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
