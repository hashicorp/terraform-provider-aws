# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_acm_acme_endpoint" "test" {
  authorization_behavior = "PRE_APPROVED"

  certificate_authority {
    public_certificate_authority {
      allowed_key_algorithms = ["RSA_2048"]
    }
  }
}

