# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Signing profiles must be in the same Region as the registry.
resource "aws_signer_signing_profile" "test" {
  region = var.region

  name        = var.rName
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_ecr_signing_configuration" "test" {
  region = var.region

  rule {
    signing_profile_arn = aws_signer_signing_profile.test.arn

    repository_filter {
      filter      = "team-a/*"
      filter_type = "WILDCARD_MATCH"
    }
  }
}

variable "rName" {
  description = "Name for the signing profile"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
