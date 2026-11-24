# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_agentregistry_registry" "test" {
  name = var.rName

  discovery_configuration {
    authorizer_type = "CUSTOM_JWT"

    authorizer_configuration {
      custom_jwt_authorizer {
        discovery_url = "https://accounts.google.com/.well-known/openid-configuration"

        custom_claim {
          inbound_token_claim_name       = "claim_99"
          inbound_token_claim_value_type = "STRING"

          authorizing_claim_match_value {
            claim_match_operator = "EQUALS"

            claim_match_value {
              match_value_string = "value01"
            }
          }
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
