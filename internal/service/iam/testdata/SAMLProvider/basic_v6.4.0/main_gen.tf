# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_saml_provider" "test" {
  name                   = var.rName
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = "https://example.com" })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.4.0"
    }
  }
}

provider "aws" {}
