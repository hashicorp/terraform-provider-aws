# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_signer_signing_profile" "test" {
  name        = var.rName
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_ecr_signing_configuration" "test" {
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
