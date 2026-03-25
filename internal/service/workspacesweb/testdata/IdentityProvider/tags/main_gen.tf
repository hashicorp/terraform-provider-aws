# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_workspacesweb_identity_provider" "test" {
  identity_provider_name = "test"
  identity_provider_type = "SAML"
  portal_arn             = aws_workspacesweb_portal.test.portal_arn

  identity_provider_details = {
    MetadataFile = file("./testfixtures/saml-metadata.xml")
  }

  tags = var.resource_tags

}

resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
