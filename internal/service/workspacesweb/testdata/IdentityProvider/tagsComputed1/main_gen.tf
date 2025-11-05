# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_workspacesweb_identity_provider" "test" {
  identity_provider_name = "test"
  identity_provider_type = "SAML"
  portal_arn             = aws_workspacesweb_portal.test.portal_arn

  identity_provider_details = {
    MetadataFile = file("./testfixtures/saml-metadata.xml")
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }

}

resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "null_resource" "test" {}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
