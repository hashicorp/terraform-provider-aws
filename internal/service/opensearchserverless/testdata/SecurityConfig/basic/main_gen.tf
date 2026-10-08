# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_security_config" "test" {
  name = var.rName
  type = "saml"
  saml_options {
    metadata = file("test-fixtures/idp-metadata.xml")
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
