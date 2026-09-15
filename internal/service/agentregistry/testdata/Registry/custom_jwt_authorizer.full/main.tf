# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_agentregistry_registry" "test" {
  name = var.rName

  discovery_configuration {
    authorizer_type = "CUSTOM_JWT"

    authorizer_configuration {
      custom_jwt_authorizer {
        discovery_url = "https://accounts.google.com/.well-known/openid-configuration"

        allowed_audience = [
          "audience-Z",
          "audience-A",
        ]
        allowed_clients = [
          "client-Z",
          "client-A",
        ]
        allowed_scopes = [
          "scope-99",
          "scope-01",
        ]

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

        custom_claim {
          inbound_token_claim_name       = "claim_01"
          inbound_token_claim_value_type = "STRING_ARRAY"

          authorizing_claim_match_value {
            claim_match_operator = "CONTAINS_ANY"

            claim_match_value {
              match_value_string_list = [
                "value99",
                "value01",
                "12345",
              ]
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
