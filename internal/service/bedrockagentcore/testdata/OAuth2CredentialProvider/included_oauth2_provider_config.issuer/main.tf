# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = var.rName

  credential_provider_vendor = "OktaOauth2"

  oauth2_provider_config {
    included_oauth2_provider_config {
      client_id     = "test-client-id"
      client_secret = "test-client-secret"

      issuer                 = "https://auth.example.com"                 # nosemgrep:ci.semgrep.domain-names.domain-names-tf
      authorization_endpoint = "https://auth.example.com/oauth2/identity" # nosemgrep:ci.semgrep.domain-names.domain-names-tf
      token_endpoint         = "https://auth.example.com/oauth2/token"    # nosemgrep:ci.semgrep.domain-names.domain-names-tf
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
