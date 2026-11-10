# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_oauth2_credential_provider" "test" {
  name = var.rName

  credential_provider_vendor = "GithubOauth2"

  oauth2_provider_config {
    github_oauth2_provider_config {
      client_id            = "test-client-id"
      client_secret_source = "EXTERNAL"

      client_secret_config {
        json_key  = "clientSecret"
        secret_id = aws_secretsmanager_secret_version.test.secret_id
      }
    }
  }
}

resource "aws_secretsmanager_secret" "test" {
  name                    = var.rName
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ clientSecret = "external-secret-value" })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
