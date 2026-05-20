# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_api_gateway_domain_name_share" "test" {
  region = var.region

  domain_name_id   = aws_api_gateway_domain_name.test.domain_name_id
  allowed_accounts = [data.aws_caller_identity.current.account_id]
}

data "aws_caller_identity" "current" {
}

resource "aws_api_gateway_domain_name" "test" {
  region = var.region

  domain_name     = var.rName
  certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["PRIVATE"]
  }
}

resource "aws_acm_certificate" "test" {
  region = var.region

  certificate_body = var.certificate_pem
  private_key      = var.private_key_pem
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
variable "certificate_pem" {
  type     = string
  nullable = false
}

variable "private_key_pem" {
  type     = string
  nullable = false
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
