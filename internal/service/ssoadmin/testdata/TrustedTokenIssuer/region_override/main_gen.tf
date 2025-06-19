# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssoadmin_trusted_token_issuer" "test" {
  region = var.region

  name                      = var.rName
  instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  trusted_token_issuer_type = "OIDC_JWT"

  trusted_token_issuer_configuration {
    oidc_jwt_configuration {
      claim_attribute_path          = "email"
      identity_store_attribute_path = "emails.value"
      issuer_url                    = "https://example.com"
      jwks_retrieval_option         = "OPEN_ID_DISCOVERY"
    }
  }
}

data "aws_ssoadmin_instances" "test" {
  region = var.region
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
