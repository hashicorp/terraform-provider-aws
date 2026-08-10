# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_secretsmanager_secret_policy" "test" {
  count = var.resource_count

  region     = var.region
  secret_arn = aws_secretsmanager_secret.test[count.index].arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "EnableAllPermissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "secretsmanager:GetSecretValue"
        Resource = "*"
      }
    ]
  })
}

resource "aws_secretsmanager_secret" "test" {
  count = var.resource_count

  region = var.region
  name   = "${var.rName}-${count.index}"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
