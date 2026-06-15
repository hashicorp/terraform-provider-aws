# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 30
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_2048"
    signing_algorithm = "SHA256WITHRSA"

    subject {
      common_name = "example.com"
    }
  }
}
