# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_security_config" "test" {
  count  = var.resource_count
  region = var.region

  name = "${substr(var.rName, 0, 30)}-${count.index}"
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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
