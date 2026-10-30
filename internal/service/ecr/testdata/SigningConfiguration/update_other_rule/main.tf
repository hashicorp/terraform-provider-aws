# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_signer_signing_profile" "a" {
  name        = "${var.rName}_a"
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "b" {
  name        = "${var.rName}_b"
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_ecr_signing_configuration" "test" {
  rule {
    signing_profile_arn = aws_signer_signing_profile.b.arn

    repository_filter {
      filter      = "*/production"
      filter_type = "WILDCARD_MATCH"
    }
  }
}

variable "rName" {
  description = "Name prefix for the signing profiles"
  type        = string
  nullable    = false
}
