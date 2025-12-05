# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_secretsmanager_secret_policy" "test" {
  region = var.region

  secret_arn = aws_secretsmanager_secret.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "EnableAllPermissions"
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action   = "secretsmanager:GetSecretValue"
      Resource = "*"
    }]
  })
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
