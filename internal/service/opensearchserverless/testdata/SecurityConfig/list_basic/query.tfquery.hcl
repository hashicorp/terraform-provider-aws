# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_opensearchserverless_security_config" "test" {
  provider = aws

  config {
    type = "saml"
  }
}
