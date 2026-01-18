# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_secretsmanager_secret_version" "test" {
  region = var.region

  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}

resource "aws_secretsmanager_secret" "test" {
  region = var.region

  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
