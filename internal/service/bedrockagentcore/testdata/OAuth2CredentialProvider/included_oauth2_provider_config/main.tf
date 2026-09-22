# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = var.rName

  credential_provider_vendor = "TwitchOauth2"

  oauth2_provider_config {
    included_oauth2_provider_config {
      client_id     = "test-client-id"
      client_secret = "test-client-secret"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
