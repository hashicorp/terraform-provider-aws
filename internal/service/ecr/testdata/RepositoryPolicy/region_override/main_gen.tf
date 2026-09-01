# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecr_repository_policy" "test" {
  region = var.region

  repository = aws_ecr_repository.test.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Sid       = var.rName
      Effect    = "Allow"
      Principal = "*"
      Action    = "ecr:ListImages"
    }]
  })
}

resource "aws_ecr_repository" "test" {
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
