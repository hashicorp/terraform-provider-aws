# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_iam_saml_provider" "test" {
  name = var.rName

  saml_metadata_document = templatefile("${path.root}/test-fixtures/saml-metadata.xml.tpl", { entity_id = "https://example.com" })

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
