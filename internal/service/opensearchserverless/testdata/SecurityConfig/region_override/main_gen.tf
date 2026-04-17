# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_security_config" "test" {
  region = var.region

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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
