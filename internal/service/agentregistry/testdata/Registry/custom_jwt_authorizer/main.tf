# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_agentregistry_registry" "test" {
  name = var.rName

  discovery_configuration {
    authorizer_type = "CUSTOM_JWT"

    authorizer_configuration {
      custom_jwt_authorizer {
        discovery_url = "https://accounts.google.com/.well-known/openid-configuration"

        allowed_clients = ["client1"]
      }
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
