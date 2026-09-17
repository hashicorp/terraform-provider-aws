# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_api_key_credential_provider" "test" {
  name                  = var.rName
  api_key_secret_source = "EXTERNAL"

  api_key_secret_config {
    secret_id = aws_secretsmanager_secret_version.test.secret_id
    json_key  = "apiKey"
  }
}

resource "aws_secretsmanager_secret" "test" {
  name = var.rName
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ apiKey = "external-secret-value" })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
