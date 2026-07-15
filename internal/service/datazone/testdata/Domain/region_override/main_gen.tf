# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_datazone_domain" "test" {
  region = var.region

  name                  = var.rName
  domain_execution_role = aws_iam_role.domain_execution_role.arn
}

resource "aws_iam_role" "domain_execution_role" {
  name = var.rName
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "datazone.amazonaws.com"
        }
      },
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "cloudformation.amazonaws.com"
        }
      },
    ]
  })

  inline_policy {
    name = var.rName
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "datazone:*",
            "ram:*",
            "sso:*",
            "kms:*",
          ]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }
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
