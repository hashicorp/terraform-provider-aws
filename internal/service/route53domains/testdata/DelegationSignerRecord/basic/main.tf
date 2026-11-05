# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53domains_delegation_signer_record" "test" {
  domain_name = var.domain_name

  signing_attributes {
    algorithm  = 13
    flags      = 257
    public_key = var.public_key
  }
}

variable "domain_name" {
  type     = string
  nullable = false
}

variable "public_key" {
  type     = string
  nullable = false
}